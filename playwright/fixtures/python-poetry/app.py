import time

running = True

try:
    while running:
        print('running')
        time.sleep(10)
except:
    running = False
