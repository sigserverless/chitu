
import os 
import json
fifoname = "hello.fifo"


data = {"key1": "value1", "len": 10}
json_data = json.dumps(data).encode()
import os
bytes_data = os.urandom(data['len'])
print(bytes_data)
print("len bytes_data ", len(bytes_data))
# Write the JSON data and the bytes data to the file
with open(fifoname, "wb") as f:
    f.write(json_data)
    f.write("\n")
    f.write(bytes_data)
    
    f.write(json_data)
    f.write("\n")
    f.write(bytes_data)
