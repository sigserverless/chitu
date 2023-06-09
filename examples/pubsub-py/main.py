import pickle 
import json 

read_fifo = open("read.fifo", 'rb')
write_fifo = open("write.fifo", 'wb')

class Op:
    def __init__(self, val):
        self.val = val

def chitu_export(key):
    req = {
        "kind": "export",
        "key": key 
    }
    write_fifo.write(json.dumps(req).encode())
    write_fifo.write(b"\n")
    write_fifo.flush()
    print("Write export", flush=True)


def chitu_import(key):
    req = {
        "kind": "import", 
        "key": key
    }
    write_fifo.write(json.dumps(req).encode())
    write_fifo.write(b"\n")
    write_fifo.flush()
    print("Write import", flush=True)


def chitu_pub(key, msg): 
    msg_bytes = pickle.dumps(msg)

    req = {
        "kind": "pub",
        "key": key, 
        "len": len(msg_bytes)
    }
    write_fifo.write(json.dumps(req).encode())
    write_fifo.write(b"\n")
    write_fifo.write(msg_bytes)
    write_fifo.flush()
    print("Write pub", flush=True)


def chitu_sub(): 
    req = read_fifo.readline()
    req = json.loads(req)
    print("Read read.fifo: {}".format(req), flush=True)
    l = req["len"]
    msg_bytes = read_fifo.read(l)
    msg = pickle.loads(msg_bytes)
    return msg 


chitu_export("x")
chitu_import("x")
chitu_pub("x", Op(1)) # You can pub anything here
op = chitu_sub()
print("Read sub: {}. (It should be like <__main__.Op object at 0xffff)".format(op), flush=True)

# clean
read_fifo.close()
write_fifo.close()