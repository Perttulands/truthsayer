// Positive cases: console.error in catch without throw

try {
  doSomething();
} catch (err) {
  console.error(err);
}

try {
  doWork();
} catch (e) {
  console.error("Error:", e.message);
  return null;
}
