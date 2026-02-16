// Negative: Express app with proper error-handling middleware
const express = require('express');
const app = express();

app.get('/users', (req, res) => {
  const users = db.getUsers();
  res.json({ users });
});

app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(3000);
