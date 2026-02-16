# Positive fixture: overly broad exception catching

try:
    process(data)
except Exception:
    handle()

try:
    connect()
except BaseException:
    cleanup()

try:
    transform(value)
except Exception as e:
    log.error(e)
