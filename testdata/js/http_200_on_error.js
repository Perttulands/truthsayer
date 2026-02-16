// Positive: res.status(200) inside catch block
const express = require('express');
const app = express();

app.get('/api/data', async (req, res) => {
  try {
    const data = await fetchData();
    res.json(data);
  } catch (err) {
    // BAD: masking error with 200 status
    res.status(200).json({ data: null });
  }
});

app.post('/api/submit', async (req, res) => {
  try {
    await save(req.body);
    res.json({ ok: true });
  } catch (err) {
    // BAD: res.json() in catch without error status
    res.json({ ok: false, error: null });
  }
});
