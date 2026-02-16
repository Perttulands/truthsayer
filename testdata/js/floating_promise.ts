// Positive: floating (unhandled) promise calls

function loadData(): void {
  fetch("/api/data");
}

function refreshCache() {
  fetch("/api/cache/refresh");
}
