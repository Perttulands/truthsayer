import subprocess

# Good: check=True raises CalledProcessError on failure
subprocess.run(["ls", "-la"], check=True)
subprocess.run(["make", "build"], capture_output=True, check=True)

# Good: check_output and check_call raise by default
subprocess.check_output(["ls"])
subprocess.check_call(["make", "build"])

# Good: explicitly handling the result
result = subprocess.run(["ls"], capture_output=True, check=True, timeout=30)
