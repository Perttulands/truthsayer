// Positive cases: rethrow without wrapping

try {
  doSomething();
} catch (err) {
  throw err;
}

try {
  doSomethingElse();
} catch (e) {
  throw e;
}

// Rethrow after some statements
try {
  doWork();
} catch (error) {
  console.log("oops");
  throw error;
}
