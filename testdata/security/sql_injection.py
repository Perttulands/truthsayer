def find_user(name, cursor):
    query = "SELECT * FROM users WHERE name = '" + name + "'"
    return cursor.execute(query)
