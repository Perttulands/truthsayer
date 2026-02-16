# Negative fixture: except with actual handling

try:
    open("file.txt")
except FileNotFoundError:
    log.warning("file not found")

try:
    connection.connect()
except ConnectionError as e:
    raise RuntimeError("connect failed") from e

try:
    parse(data)
except ValueError:
    default_value = get_default()
    return default_value
