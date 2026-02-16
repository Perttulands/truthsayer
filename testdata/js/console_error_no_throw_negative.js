// Negative cases: console.error with throw, or no console.error

try {
  doSomething();
} catch (err) {
  console.error(err);
  throw err;
}

try {
  doWork();
} catch (e) {
  console.error("Error:", e.message);
  throw new Error("failed", { cause: e });
}

// No console.error
try {
  parse();
} catch (err) {
  throw err;
}

// console.log (not console.error)
try {
  load();
} catch (err) {
  console.log(err);
}

// Empty catch
try {
  optional();
} catch (err) {
}
