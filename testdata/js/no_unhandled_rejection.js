// Positive: entry point without unhandledRejection handler
const express = require('express');
const app = express();

app.use(bodyParser.json());
app.use('/api', routes);

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
