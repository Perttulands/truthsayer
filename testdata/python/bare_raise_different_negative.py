# Chained with from — correct
try:
    connect()
except ConnectionError as e:
    raise RuntimeError("failed") from e

# Bare raise — re-raises same exception, correct
try:
    parse()
except ValueError:
    raise

# Raise same exception type (still caught by this rule for now, but it's re-raising)
try:
    load()
except KeyError as e:
    raise KeyError("better message") from e
