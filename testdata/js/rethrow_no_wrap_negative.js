// Negative cases: proper error wrapping

try {
  doSomething();
} catch (err) {
  throw new Error("failed to do something", { cause: err });
}

try {
  doSomethingElse();
} catch (e) {
  throw new CustomError("context", e);
}

// Throw a different error
try {
  parse();
} catch (err) {
  throw new TypeError("invalid input");
}

// No throw in catch
try {
  load();
} catch (err) {
  console.error(err);
  return null;
}

// Empty catch with comment
try {
  optional();
} catch (err) {
  // intentionally ignored
}
