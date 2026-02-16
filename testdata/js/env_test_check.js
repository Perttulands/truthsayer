// Positive: NODE_ENV test check in production code
if (process.env.NODE_ENV === 'test') {
  db = mockDatabase;
} else {
  db = realDatabase;
}

const isTest = 'test' === process.env.NODE_ENV;
