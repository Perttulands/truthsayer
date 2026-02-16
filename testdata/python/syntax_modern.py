# Modern Python syntax for tree-sitter parse validation

# Type hints (3.5+)
def greet(name: str) -> str:
    return f"Hello, {name}"

# f-strings (3.6+)
value = 42
message = f"The answer is {value}"
nested = f"{'yes' if value > 0 else 'no'}"

# Variable annotations (3.6+)
count: int = 0
names: list[str] = []

# Dataclasses (3.7+)
from dataclasses import dataclass

@dataclass
class Point:
    x: float
    y: float
    label: str = "origin"

# Walrus operator (3.8+)
import re
if match := re.search(r"\d+", "abc123"):
    print(match.group())

data = [1, 2, 3, 4, 5]
if (n := len(data)) > 3:
    print(f"List has {n} elements")

# Positional-only params (3.8+)
def pos_only(a, b, /, c, d):
    return a + b + c + d

# Dictionary merge operators (3.9+)
defaults = {"color": "red", "size": 10}
overrides = {"size": 20, "weight": "bold"}
merged = defaults | overrides

# Type union syntax (3.10+)
def process(value: int | str) -> str:
    return str(value)

# Match/case (3.10+)
def handle_command(command):
    match command:
        case "quit":
            return False
        case "hello":
            return True
        case ["go", direction]:
            return f"Going {direction}"
        case {"action": action, "target": target}:
            return f"{action} -> {target}"
        case _:
            return None

# Exception groups (3.11+)
# try:
#     raise ExceptionGroup("errors", [ValueError("a"), TypeError("b")])
# except* ValueError as eg:
#     print(eg.exceptions)

# Async/await
import asyncio

async def fetch_data(url: str) -> dict:
    await asyncio.sleep(0.1)
    return {"url": url, "status": 200}

async def main():
    tasks = [fetch_data(f"http://example.com/{i}") for i in range(5)]
    results = await asyncio.gather(*tasks)
    return results

# Generators
def fibonacci():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

# Context managers
class ManagedResource:
    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        return False

# Decorators with arguments
def retry(times: int = 3):
    def decorator(func):
        def wrapper(*args, **kwargs):
            for i in range(times):
                try:
                    return func(*args, **kwargs)
                except Exception:
                    if i == times - 1:
                        raise
        return wrapper
    return decorator

@retry(times=5)
def unstable_operation():
    pass

# Star expressions in assignments
first, *middle, last = [1, 2, 3, 4, 5]

# Multiple inheritance
class A:
    pass

class B:
    pass

class C(A, B):
    pass

# Lambda expressions
square = lambda x: x ** 2

# Comprehensions
squares = [x**2 for x in range(10)]
evens = {x for x in range(20) if x % 2 == 0}
mapping = {k: v for k, v in zip("abc", [1, 2, 3])}
gen = (x**2 for x in range(10))
