// Positive fixture: callback error parameter ignored

const fs = require('fs');

// Bad: err parameter never used
fs.readFile('config.json', (err, data) => {
  console.log(data);
});

// Bad: error parameter never used
doWork((error, result) => {
  return result;
});

// Bad: function expression callback
fs.stat('file.txt', function(err, stats) {
  return stats.size;
});
