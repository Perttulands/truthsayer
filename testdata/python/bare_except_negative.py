# Negative fixture: proper except clauses with types

try:
    open("file.txt")
except FileNotFoundError:
    handle()

try:
    connection.connect()
except ConnectionError as e:
    log.error(f"connection failed: {e}")

try:
    parse(data)
except (ValueError, TypeError):
    handle_parse_error()

try:
    something()
except Exception as e:
    raise RuntimeError("wrapped") from e
