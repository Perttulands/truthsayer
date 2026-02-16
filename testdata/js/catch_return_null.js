// Positive: catch handlers returning null/undefined

const data = fetch("/api").catch(() => null);

const result = promise.catch(() => undefined);

const value = doWork().catch((err) => { return null; });
