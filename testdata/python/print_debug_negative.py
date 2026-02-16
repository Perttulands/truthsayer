# Proper logging usage
import logging

logger = logging.getLogger(__name__)

def process_data(data):
    logger.info("processing...")
    result = transform(data)
    logger.debug("result: %s", result)
    return result

# print in comments is fine
# print("this is a comment")
