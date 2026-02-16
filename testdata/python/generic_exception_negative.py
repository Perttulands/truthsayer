raise ValueError("invalid value")

raise TypeError("expected string")

raise RuntimeError("connection lost")

# Bare raise is fine
try:
    x()
except Exception:
    raise

class CustomError(Exception):
    pass

raise CustomError("custom message")
