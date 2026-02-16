data = {"name": "Alice", "age": 30}

# Good: explicit default provided
name = data.get("name", "unknown")
age = data.get("age", 0)

# Good: even None is explicit — developer made a conscious choice
debug = data.get("debug", None)

# Good: direct access raises KeyError
email = data["email"]
