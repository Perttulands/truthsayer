# Positive: requests without status checking
import requests

response = requests.get("https://api.example.com/data")
data = response.json()

requests.post("https://api.example.com/submit", json={"key": "value"})
