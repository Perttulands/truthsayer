# Good: immutable module-level constants
MAX_RETRIES = 3
API_URL = "https://api.example.com"
VALID_STATUSES = (200, 201, 204)
DEBUG = False
VERSION = "1.0.0"

# Class-level state is fine (encapsulated)
class Registry:
    _handlers = {}

# Function-level state is fine
def get_cache():
    cache = {}
    return cache
