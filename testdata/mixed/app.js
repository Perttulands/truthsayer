const express = require("express");
const app = express();

// Anti-pattern: empty catch
try {
    doSomething();
} catch (e) {}

// Anti-pattern: eval usage
eval("console.log('hello')");

app.listen(3000);
