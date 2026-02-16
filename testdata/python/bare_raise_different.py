try:
    connect()
except ConnectionError:
    raise RuntimeError("failed")

try:
    parse()
except ValueError:
    raise TypeError("bad type")

try:
    load()
except KeyError as e:
    raise ValueError("missing key")
