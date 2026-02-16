# Positive: debug flags with side effects in production code
import logging

if __debug__:
    print("Debug mode is on")

if DEBUG:
    logging.info("debugging enabled")

x = 10
if __debug__:
    x = 20  # assignment only, but it's an expression_statement... actually this IS a side effect
