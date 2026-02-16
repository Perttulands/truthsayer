import logging

try:
    connect()
except ConnectionError as e:
    logging.error(e)
    raise

try:
    parse()
except ValueError as e:
    logger.exception(e)
    raise

try:
    load()
except Exception as e:
    log.critical(e)
    raise RuntimeError("failed")
