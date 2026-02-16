try:
    connect()
except ConnectionError as e:
    raise RuntimeError("connection failed") from None

try:
    parse()
except ValueError as e:
    raise TypeError("bad data") from None
