document.addEventListener("htmx:beforeSwap", function (event) {
  const xhr = event.detail.xhr;
  const contentType = xhr.getResponseHeader("Content-Type") || "";
  if (xhr.status >= 400 && xhr.status < 600 && contentType.includes("text/html")) {
    event.detail.shouldSwap = true;
    event.detail.isError = false;
  }
});
