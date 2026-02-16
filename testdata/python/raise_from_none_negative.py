try:
    connect()
except ConnectionError as e:
    raise RuntimeError("connection failed") from e

try:
    parse()
except ValueError as e:
    raise TypeError("bad data") from e

# bare raise is fine
try:
    x()
except Exception:
    raise

# raise without from inside except is caught by bare-raise-different, not this rule
try:
    x()
except ValueError:
    raise TypeError("oops")
