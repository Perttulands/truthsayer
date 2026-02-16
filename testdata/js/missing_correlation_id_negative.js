// Negative: Express app with request ID middleware
const express = require('express');
const app = express();

app.use((req, res, next) => {
  req.id = req.headers['x-request-id'] || crypto.randomUUID();
  next();
});

app.get('/api/users', (req, res) => {
  res.json({ users: [] });
});

app.listen(3000);
