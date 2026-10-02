// The server always returns a full document. Select #page for both successful
// navigation and error responses, keeping the browser's native form fallback.
document.addEventListener("htmx:beforeSwap", function (event) {
  if (event.detail.xhr.status === 400) {
    event.detail.shouldSwap = true;
    event.detail.isError = false;
  }
});
