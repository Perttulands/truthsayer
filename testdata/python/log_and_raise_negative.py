import logging

# Only logging, no raise — fine (handled elsewhere)
try:
    connect()
except ConnectionError as e:
    logging.error(e)
    return None

# Only raise, no logging — fine
try:
    parse()
except ValueError as e:
    raise

# Logging info level + raise — not error logging
try:
    load()
except Exception as e:
    logging.info("retrying after error")
    raise

# No logging at all
try:
    x()
except Exception as e:
    print(e)
    raise
