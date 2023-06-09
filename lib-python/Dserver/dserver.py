import os
# from Dserver.ddict import DDict
from Dserver.pdict import DDict 
from Dserver.constants import PIPE_DICT
import json


class DServer:

    def __init__(self, inv_id = os.environ.get("INV_ID"), dag_id = os.environ.get("DAG_ID")) -> None:
        self.DMap = {}  # key-did val-Tensor
        self.inv_id = inv_id
        self.dag_id = dag_id
        self.opened_files = {}


    def new_D(self, dname, type=2):  # 分pipe
        handle_msg = json.dumps({
            "DType": type
        })
        self.write_pipe(handle_msg, PIPE_DICT["cons_req"])

        res = self.read_pipe(PIPE_DICT['cons_res'])
        did = json.loads(res)['Id']
        new_D = DDict(did, self.opened_files)
        self.DMap[did] = new_D
        return new_D


    def import_D(self, dname):
        handle_msg = json.dumps({
            "Key":   dname,
            "InvId": self.inv_id,
            "DagId": self.dag_id,
            "DType": 2
        })
        self.write_pipe(handle_msg, PIPE_DICT['import_req'])
        res = self.read_pipe(PIPE_DICT['import_res']) 
        did = json.loads(res)['Id']

        imported_D = DDict(did, self.opened_files)
        self.DMap[did] = imported_D
        return imported_D
       

    def export_D(self, ddict, dname):
        handle_msg = json.dumps({
            "Id" : ddict.did,
            "InvId": self.inv_id,
            "DagId": self.dag_id,
            "Key": dname, 
        })
        self.write_pipe(handle_msg, PIPE_DICT['export_req'])


    def trigger_D(self, fname, args): 
        handle_msg = json.dumps({
            "Fname": fname, 
            "DagId": self.dag_id, 
            "Args": args,
        })
        self.write_pipe(handle_msg, PIPE_DICT['trigger_req'])


    def merge_Ds(self, ddict1, ddict2):
        handle_msg = json.dumps({
                "Ids": [ddict1.did, ddict2.did],
                "Rop": {
                    "kind" : "MergeMean",
                    "val" : {}
                }            
            })
        self.write_pipe(handle_msg, PIPE_DICT['read_req'])
        res = self.read_pipe(PIPE_DICT['read_res'])
        did = json.loads(res)['ReaderId']
        merged_D = DDict(did, self.opened_files)
        self.DMap[did] = merged_D
        return merged_D
    
    def merge_Ds(self, did_list):
        handle_msg = json.dumps({
                "Ids": [did for did in did_list],
                "Rop": {
                    "kind" : "MergeMean",
                    "val" : {}
                }            
            })
        self.write_pipe(handle_msg, PIPE_DICT['read_req'])
        res = self.read_pipe(PIPE_DICT['read_res'])
        did = json.loads(res)['ReaderId']
        merged_D = DDict(did, self.opened_files)
        self.DMap[did] = merged_D
        return merged_D


    def write_pipe(self, content, fifoname):
        if self.opened_files.get(fifoname):
            fifo = self.opened_files[fifoname]
        else: 
            fifo = open(fifoname, 'w')
            self.opened_files[fifoname] = fifo 

        fifo.write(content+"\n")
        fifo.flush()

    
    def read_pipe(self, fifoname):
        if self.opened_files.get(fifoname):
            fifo = self.opened_files[fifoname]
        else: 
            fifo = open(fifoname, 'r')
            self.opened_files[fifoname] = fifo 

        while True: 
            res = fifo.readline()
            if len(res) != 0: 
                return res 
            else: 
                fifo.close() 
                fifo = open(fifoname, 'r')
                self.opened_files[fifoname] = fifo 


    def close(self): 
        for fifo in self.opened_files.values():
            fifo.close()
