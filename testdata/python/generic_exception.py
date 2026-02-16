raise Exception("something failed")

raise BaseException("critical")

def process():
    if not valid:
        raise Exception("invalid data")
