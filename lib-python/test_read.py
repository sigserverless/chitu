

import os 
import json
fifoname = "hello.fifo"
if not os.path.exists(fifoname):
    os.mkfifo(fifoname)  
# Read the JSON data and the bytes data from the file
with open(fifoname, "rb") as f:
    json_data = f.readline().strip()  # Remove the trailing newline
    bytes_data = f.read(10)
    data = json.loads(json_data.decode())
    print(data['key1']) 
    json_data = f.readline().strip()  # Remove the trailing newline
    bytes_data = f.read(10)
    data = json.loads(json_data.decode())
    print(data['key1']) 
