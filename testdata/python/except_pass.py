# Positive fixture: except with only pass

try:
    open("file.txt")
except FileNotFoundError:
    pass

try:
    connection.connect()
except ConnectionError:
    pass

try:
    parse(data)
except:
    pass
