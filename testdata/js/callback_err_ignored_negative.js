// Negative fixture: callback error parameter properly handled

const fs = require('fs');

// Good: err is checked
fs.readFile('config.json', (err, data) => {
  if (err) throw err;
  console.log(data);
});

// Good: error is logged
doWork((error, result) => {
  if (error) {
    console.error(error);
    return;
  }
  return result;
});

// Good: single parameter (not err-first pattern)
items.forEach((item) => {
  console.log(item);
});

// Good: non-err first param name
doWork((response, data) => {
  console.log(data);
});

// Good: not a callback (standalone function)
const handler = (err, data) => {
  console.log(data);
};
