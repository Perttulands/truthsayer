// Negative cases: descriptive error messages

throw new Error("Failed to parse config file: " + path);
throw new Error(`User ${userId} not found in database`);
throw new TypeError("Expected a number but got a string");
throw new Error("Connection to redis timed out after 30s");
throw new Error("Invalid email format for registration");

// Not an Error constructor
throw new CustomError("failed");

// No arguments
throw new Error();
