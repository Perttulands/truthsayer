def find_user(name, cursor):
    query = "SELECT * FROM users WHERE name = %s"
    return cursor.execute(query, (name,))
