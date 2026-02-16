# Good: immutable defaults and None pattern
def append_to(items=None):
    if items is None:
        items = []
    items.append(1)
    return items

def make_config(options=None):
    if options is None:
        options = {}
    return options

def greet(name="world"):
    return f"Hello, {name}!"

def count(start=0, end=10):
    return range(start, end)

def toggle(flag=True):
    return not flag

def choose(value=(1, 2, 3)):
    return value
