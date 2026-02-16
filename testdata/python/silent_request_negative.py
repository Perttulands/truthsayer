# Negative: requests with proper status checking
import requests

# Using raise_for_status
response = requests.get("https://api.example.com/data")
response.raise_for_status()

# Using status_code check
resp = requests.post("https://api.example.com/submit", json={"key": "value"})
if resp.status_code != 200:
    raise Exception("Request failed")

# Chained call
requests.get("https://api.example.com/health").raise_for_status()

# Using .ok property
r = requests.get("https://api.example.com/check")
if not r.ok:
    raise Exception("Not ok")
