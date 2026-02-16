# Bad: requests calls without timeout
import requests

response = requests.get("https://api.example.com/data")
response = requests.post("https://api.example.com/submit", json={"key": "value"})
response = requests.put("https://api.example.com/update", data="payload")
response = requests.delete("https://api.example.com/item/1")
response = requests.patch("https://api.example.com/item/1", json={"status": "done"})
