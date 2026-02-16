import subprocess

def run_task(cmd):
    # Anti-pattern: bare except
    try:
        result = subprocess.run(cmd)
    except:
        pass

def load_items(items=[]):
    return items
