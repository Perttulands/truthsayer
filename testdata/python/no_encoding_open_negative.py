# Good: open() with encoding or binary mode
f = open("data.txt", encoding="utf-8")
f = open("data.txt", "r", encoding="utf-8")

# Binary mode — encoding not needed
f = open("image.png", "rb")
f = open("output.bin", "wb")
f = open("data.dat", mode="rb")

with open("config.json", encoding="utf-8") as cfg:
    data = cfg.read()
