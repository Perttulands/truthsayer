raise Exception("something went wrong")

raise ValueError("error occurred")

# Bare raise is fine
try:
    x()
except Exception:
    raise
