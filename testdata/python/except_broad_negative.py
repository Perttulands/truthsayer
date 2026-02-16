# Negative fixture: specific exception catching

try:
    process(data)
except ValueError:
    handle()

try:
    connect()
except ConnectionError as e:
    raise RuntimeError("connect failed") from e

try:
    transform(value)
except (KeyError, IndexError):
    return default

try:
    something()
except KeyError as e:
    log.warning(f"missing key: {e}")
