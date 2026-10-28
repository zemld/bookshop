import http.server
import pathlib
import subprocess
import threading
import time
import unittest
import urllib.parse
import urllib.request


class ServeCatalog(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.server.requests.append(self.path)
        if self.path == "/slow":
            time.sleep(4)
        status = {"/error": 500, "/redirect": 302}.get(self.path, 200)
        self.send_response(status)
        if status == 302:
            self.send_header("Location", "/ok")
        self.end_headers()
        try:
            self.wfile.write(b"arbitrary response")
        except BrokenPipeError:
            pass

    def log_message(self, format, *args):
        pass


class CheckCatalogProbe(unittest.TestCase):
    def test_probe(self):
        config = pathlib.Path(__file__).parent / "blackbox" / "blackbox.yml"
        self.assertTrue(config.is_file(), "missing exporter configuration")
        server = http.server.ThreadingHTTPServer(("0.0.0.0", 0), ServeCatalog)
        server.requests = []
        threading.Thread(target=server.serve_forever, daemon=True).start()
        container = None
        try:
            container = subprocess.check_output(
                [
                    "docker", "run", "--rm", "-d",
                    "--add-host=host.docker.internal:host-gateway",
                    "-p", "127.0.0.1::9115",
                    "-v", f"{config.resolve()}:/config/blackbox.yml:ro",
                    "quay.io/prometheus/blackbox-exporter:v0.27.0",
                    "--config.file=/config/blackbox.yml",
                ],
                text=True,
            ).strip()
            address = subprocess.check_output(
                ["docker", "port", container, "9115/tcp"], text=True
            ).strip()
            for attempt in range(50):
                try:
                    with urllib.request.urlopen(f"http://{address}/metrics", timeout=1):
                        break
                except OSError:
                    if attempt == 49:
                        raise
                    time.sleep(0.1)
            for path, expected in [("/ok", 1), ("/error", 0), ("/redirect", 0), ("/slow", 0)]:
                with self.subTest(path=path):
                    target = f"http://host.docker.internal:{server.server_port}{path}"
                    query = urllib.parse.urlencode({"module": "catalog", "target": target})
                    started = time.monotonic()
                    with urllib.request.urlopen(
                        f"http://{address}/probe?{query}", timeout=5
                    ) as response:
                        metrics = response.read().decode()
                    self.assertIn(f"\nprobe_success {expected}\n", metrics)
                    self.assertLess(time.monotonic() - started, 5)
                    if path == "/redirect":
                        self.assertIn("\nprobe_http_status_code 302\n", metrics)
            self.assertEqual(server.requests, ["/ok", "/error", "/redirect", "/slow"])
        finally:
            if container:
                subprocess.run(["docker", "stop", container], check=True, stdout=subprocess.DEVNULL)
            server.shutdown()
            server.server_close()


if __name__ == "__main__":
    unittest.main()
