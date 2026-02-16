# Entry point with proper logging config
import sys
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def main():
    logger.info("starting")
    data = sys.argv[1]
    process(data)

if __name__ == "__main__":
    main()
