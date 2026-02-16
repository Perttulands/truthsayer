// Negative: catch handlers with proper handling

const data = fetch("/api").catch((err) => {
  console.error(err);
  return defaultValue;
});

const result = promise.catch((err) => {
  throw new Error(`Failed: ${err.message}`);
});

const value = doWork().catch((err) => {
  reportError(err);
  return fallbackResult;
});
