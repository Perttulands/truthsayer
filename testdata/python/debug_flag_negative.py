# Negative: debug flags without side effects (just pass)
if __debug__:
    pass

# Regular if statements
if x > 0:
    print("positive")

# DEBUG as part of a longer name (should not trigger)
if DEBUG_MODE:
    print("debug mode")
