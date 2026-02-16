# Bad: getattr with None default masks missing attributes
x = getattr(obj, 'name', None)
y = getattr(config, 'debug_mode', None)
