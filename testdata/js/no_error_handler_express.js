// Positive: Express app with routes but no error-handling middleware
const express = require('express');
const app = express();

app.get('/users', (req, res) => {
  const users = db.getUsers();
  res.json({ users });
});

app.post('/users', (req, res) => {
  const user = db.createUser(req.body);
  res.status(201).json(user);
});

app.listen(3000);
