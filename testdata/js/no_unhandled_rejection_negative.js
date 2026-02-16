// Negative: entry point with unhandledRejection handler
const express = require('express');
const app = express();

process.on('unhandledRejection', (err) => {
  console.error('Unhandled rejection:', err);
  process.exit(1);
});

app.use(bodyParser.json());
app.listen(3000);
