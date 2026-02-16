function loadData() {
  try {
    runTask();
  } catch (err) {
    console.error(err);
    throw err;
  }
}
