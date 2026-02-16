# Bad: mutable default arguments
def append_to(items=[]):
    items.append(1)
    return items

def make_config(options={}):
    options["key"] = "value"
    return options

def collect(values=set()):
    values.add(42)
    return values

class MyClass:
    def method(self, data=[1, 2, 3]):
        return data
