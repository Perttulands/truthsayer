// Negative: proper error status in catch blocks
const express = require('express');
const app = express();

app.get('/api/data', async (req, res) => {
  try {
    const data = await fetchData();
    res.json(data);
  } catch (err) {
    // GOOD: proper error status
    res.status(500).json({ error: err.message });
  }
});

app.post('/api/submit', async (req, res) => {
  try {
    await save(req.body);
    res.status(201).json({ ok: true });
  } catch (err) {
    // GOOD: 400 error status
    res.status(400).json({ error: 'Bad request' });
  }
});

// GOOD: status 200 outside catch block
app.get('/api/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});
