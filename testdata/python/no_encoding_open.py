# Bad: open() without encoding
f = open("data.txt")
f = open("data.txt", "r")
f = open("data.txt", "w")

with open("config.json") as cfg:
    data = cfg.read()

with open("output.csv", "w") as out:
    out.write("hello")
