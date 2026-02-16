// Positive: console.log in production code
export function handleRequest(req: Request) {
  console.log(req.body);
  console.log('Processing request');
  return processData(req);
}
