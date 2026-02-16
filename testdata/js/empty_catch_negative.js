// Negative: catch blocks with proper handling — should NOT be flagged

try {
  riskyOperation();
} catch (e) {
  console.error(e);
}

try {
  anotherOp();
} catch (err) {
  // intentionally ignored — operation is optional
}

try {
  moreStuff();
} catch (e) {
  logError(e);
  throw e;
}
