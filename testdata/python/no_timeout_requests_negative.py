# Good: requests calls with timeout
import requests

response = requests.get("https://api.example.com/data", timeout=30)
response = requests.post("https://api.example.com/submit", json={"key": "value"}, timeout=10)
response = requests.put("https://api.example.com/update", timeout=(5, 30))

# Using **kwargs (might contain timeout)
config = {"timeout": 30}
response = requests.get("https://api.example.com/data", **config)
