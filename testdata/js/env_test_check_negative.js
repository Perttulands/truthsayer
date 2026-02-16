// Negative: Non-test NODE_ENV checks are fine
if (process.env.NODE_ENV === 'production') {
  enableCaching();
}

if (process.env.NODE_ENV === 'development') {
  enableDevTools();
}

// Other env vars are fine
if (process.env.DEBUG === 'true') {
  enableLogging();
}
