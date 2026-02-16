# Production code using print for debugging
import os

def process_data(data):
    print("processing...")
    result = transform(data)
    print(f"result: {result}")
    return result

def handle_error(err):
    print(err)
    return None
