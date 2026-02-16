// Positive: empty catch blocks that should be flagged

try {
  riskyOperation();
} catch (e) {
}

try {
  anotherOp();
} catch (err) {
}

try {
  moreStuff();
} catch {
}
