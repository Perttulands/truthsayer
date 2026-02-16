# Production code with proper credential handling
import os

password = os.environ["DB_PASSWORD"]
api_key = os.environ.get("API_KEY")
secret = ""  # empty string is fine

# Commented-out credentials are fine
# password = "test"
