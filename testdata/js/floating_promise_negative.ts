// Negative: properly handled promise calls

async function loadData() {
  const data = await fetch("/api/data");
  return data;
}

function loadWithThen() {
  fetch("/api/data").then(handleResponse);
}

function loadWithCatch() {
  fetch("/api/data").catch(handleError);
}

function loadAndReturn() {
  return fetch("/api/data");
}

const promise = fetch("/api/data");
