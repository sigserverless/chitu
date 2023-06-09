def main():
    import sys
    import time 
    s = sys.stdin.readline().strip()

    for i in range(5):
        print("Hello: " + s + " " + str(i), flush=True)
        time.sleep(1)

main()