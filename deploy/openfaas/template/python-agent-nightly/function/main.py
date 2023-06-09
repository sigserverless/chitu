print("Subprocess executed.", flush=True)

read_fifo = open("read.fifo", 'r')
write_fifo = open("write.fifo", 'w')

import os 

print("Hello, dag: " + os.environ["DAG_ID"], flush=True)

read_fifo.close()
write_fifo.close()