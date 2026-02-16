data = {"name": "Alice", "age": 30}

# Bad: dict.get without explicit default returns None
name = data.get("name")
age = data.get("age")
