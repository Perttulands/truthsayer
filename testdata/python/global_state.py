# Bad: module-level mutable globals
ITEMS = []
CACHE = {}
SEEN = set()

# These are also bad
registry = []
handlers = {}
