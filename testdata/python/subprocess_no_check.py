import subprocess

# Bad: no check=True
subprocess.run(["ls", "-la"])
subprocess.call(["make", "build"])
subprocess.run(["deploy.sh"], capture_output=True)
