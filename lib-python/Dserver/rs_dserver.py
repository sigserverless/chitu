import os
from Dserver.rs_ddict import DDict 
import json
import queue
import threading
import pickle
import requests

def write_thread(q, fifo):
    while True:
        dct = q.get()  
        fifo.write(json.dumps(dct).encode())
        fifo.write(b"\n")
        if "len" in dct and dct['len'] != 0: 
            msg = q.get()  
            fifo.write(msg)
        fifo.flush()


def read_thread(q, fifo):
    
    while True:
        # print("Read sub from read.fifo ", flush=True)
        data = fifo.readline()
        # print("Read sub from read.fifo after readline ", flush=True)
        data = json.loads(data)
        len = data['len']
        if len == 0:
            q.put(("end", "end"))
        else :
            msg = fifo.read(len)
            msg = pickle.loads(msg)
            q.put(msg)
        


class DServer:

    def __init__(self, inv_id = os.environ.get("INV_ID"), dag_id = os.environ.get("DAG_ID")) -> None:
        self.DMap = {}  
        self.inv_id = inv_id
        self.dag_id = dag_id
        # launch the write-pipe and read-pipe threads
        self.write_q = queue.Queue()
        self.read_q = queue.Queue()
        self.write_fifo = open('write.fifo', "wb") 
        self.read_fifo = open('read.fifo', "rb") 
        self.writer = threading.Thread(target=write_thread, args=(self.write_q, self.write_fifo))
        self.reader = threading.Thread(target=read_thread, args=(self.read_q, self.read_fifo))
        self.writer.start()
        self.reader.start()

    def new_D(self, dname): 
        new_D = DDict(dname, self.dag_id, self.write_q, self.read_q)
        return new_D

    def import_D(self, dname, MERGED):
        data = {
            "kind" : "import",
            "key" : dname
        }
        self.write_q.put(data)
        
        ddict = DDict(dname, self.dag_id, self.write_q, self.read_q)
        ddict.IMPORT = True
        ddict.MERGED = MERGED
        ddict.read_fifo() 
        return ddict
       

    def export_D(self, ddict, dname):
        data = {
            "kind" : "export",
            "key" : dname
        }
        self.write_q.put(data)
        ddict.EXPORT = True
        return ddict

    def trigger_F(self, fname, args = ""):
        url = 'http://coordinator.openfaas-fn:8080/trigger' 
        data = {
            "dagId": self.dag_id,
            "fname": fname,
            "args": args
        }  
        response = requests.post(url, json=data)
        print("send trigger request ", flush=True)

    def merge_Ds(self, dname, ddicts): 
        # work only for worker1 
        merged_D = DDict(dname, self.dag_id, self.write_q, self.read_q)
        for dct in ddicts:
            merged_D.register_watchee(dct)
            dct.register_observer(merged_D)
            if dct.total_keys != -1:
                merged_D.total_keys = dct.total_keys
        return merged_D

    def close(self):
        self.write_fifo.close()
        self.read_fifo.close()