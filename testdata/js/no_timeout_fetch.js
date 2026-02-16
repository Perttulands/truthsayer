// Bad: fetch without abort signal
async function loadData() {
  const response = await fetch("/api/users");
  const data = await fetch("/api/posts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  });
  return data;
}
