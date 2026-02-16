# Positive fixture: bare except clauses

try:
    open("file.txt")
except:
    pass

try:
    connection.connect()
except:
    print("connection failed")

try:
    parse(data)
except:
    return None
