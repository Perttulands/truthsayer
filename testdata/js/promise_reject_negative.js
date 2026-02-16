// Negative cases: reject with Error objects

Promise.reject(new Error("something failed"));
Promise.reject(new TypeError("invalid input"));

new Promise((resolve, reject) => {
  reject(new Error("bad request"));
});

// Variable (could be an Error)
Promise.reject(err);
Promise.reject(error);

// No arguments
Promise.reject();
