// Good: fetch with abort signal
async function loadData() {
  const controller = new AbortController();
  setTimeout(() => controller.abort(), 5000);
  const response = await fetch("/api/users", { signal: controller.signal });
  const data = await fetch("/api/posts", { signal: AbortSignal.timeout(5000) });
  return data;
}
