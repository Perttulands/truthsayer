# Good: explicit non-None default
x = getattr(obj, 'name', 'default')

# Good: 2-arg getattr raises AttributeError
y = getattr(obj, 'name')

# Good: explicit non-None defaults
z = getattr(obj, 'items', [])
w = getattr(obj, 'enabled', False)
