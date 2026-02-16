// Positive cases: reject with non-Error values

Promise.reject("something failed");
Promise.reject(42);
Promise.reject(null);
Promise.reject(undefined);
Promise.reject(true);

new Promise((resolve, reject) => {
  reject("bad request");
});

new Promise((resolve, reject) => {
  reject(404);
});
