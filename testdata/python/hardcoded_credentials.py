# Production code with hardcoded credentials
import requests

password = "supersecret123"
api_key = "sk-1234567890abcdef"
secret = "my-secret-value"
token = "bearer-abc-xyz"

def connect():
    return requests.post(url, auth=("user", password))
