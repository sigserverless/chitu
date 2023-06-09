from collections import OrderedDict
import os
import errno
import json
from Dserver.constants import PIPE_DICT
import multiprocessing
import Dserver.pdict_pb2 as pdict_pb2 
import struct
import time


def send_layer_grad_func(did, key, val, write_req_fifo_name, write_res_fifo_name):
    msg = pdict_pb2.PDictSet()
    msg.id = did
    msg.key = key
    msg.values.values.extend(val)
    data = msg.SerializeToString()
    data_size = len(data)
    print("data size: {}, key: {}".format(data_size, key), flush=True)
    # write
    write_start_time = time.time()
    write_req_fifo = open(write_req_fifo_name, 'wb')
    write_req_fifo.write(struct.pack('>I', data_size))
    write_req_fifo.write(data)
    write_req_fifo.flush()
    write_end_time = time.time()
    write_fifo_time = write_end_time - write_start_time
    # read
    write_res_fifo = open(write_res_fifo_name, 'r')
    while True: 
        res = write_res_fifo.readline()
        if len(res) != 0: 
            read_fifo_time = time.time() - write_end_time
            return write_fifo_time, read_fifo_time

class DDict:
    
    def __init__(self, did, opened_files) -> None:
        self.data_dict = OrderedDict()
        self.did = did
        self.total_layers = 0
        self.curr_layer_count = 0
        self.hooks = []
        self.opened_files = opened_files
        self.set_time = 0
        self.write_fifo_time = 0
        self.read_fifo_time = 0
        # TODO: 
        self.processes = []
        self.process_pool = multiprocessing.Pool(processes = 1)

    
    def register_params(self, params_iterator):
        self.managed_param_dict = {key : val for key, val in params_iterator}
        self.total_layers = len(self.managed_param_dict) 
        self.gen_idx_to_key_dict()
        def hook_func(grad):
            # time start
            set_start_time = time.time()
            self.set(self.idx_to_layer_name[self.curr_layer_count], grad.reshape(-1).tolist())
            # self.set(self.idx_to_layer_name[self.curr_layer_count], grad)
            self.curr_layer_count += 1
            self.set_time += (time.time() - set_start_time)
         
        for _, val in self.managed_param_dict.items():
            hook = val.register_hook(hook_func)
            self.hooks.append(hook)
        
    
    def end(self):
        # Wait for all processes to finish
        # for process in self.processes:
        #     process.join()
        self.process_pool.close()    
        self.process_pool.join()
        
        handle_msg = json.dumps({
            "Id" : self.did  
        })
        self.write_pipe(handle_msg, PIPE_DICT['end_req'])
        self.reset()

    
    def reset(self):
        self.curr_layer_count = 0
        self.data_dict = OrderedDict()
        for hook in self.hooks:
            hook.remove()
        self.processes.clear()

    
    def gen_idx_to_key_dict(self):
        
        idx_to_layer_name = OrderedDict()
        count = 0
        for key, _ in self.managed_param_dict.items():
            idx_to_layer_name[count] = key
            count += 1

        for i in range(self.total_layers):
            if i != (self.total_layers-1):
                if idx_to_layer_name[i].endswith("weight") and idx_to_layer_name[i+1].endswith("bias") and not idx_to_layer_name[i].startswith("fc"):
                    tmp = idx_to_layer_name[i]
                    idx_to_layer_name[i] = idx_to_layer_name[i+1]
                    idx_to_layer_name[i+1] = tmp
                    i += 1
        layer_name_to_idx = {}
        for k, v in idx_to_layer_name.items():
            layer_name_to_idx[v] = self.total_layers -1 - k
        for k, v in layer_name_to_idx.items():
            idx_to_layer_name[v] = k 
            
        self.idx_to_layer_name = idx_to_layer_name
        self.layer_name_to_idx = layer_name_to_idx

    def wait(self):
        handle_msg = json.dumps({
            "Id" : self.did  
        })
        self.write_pipe(handle_msg, PIPE_DICT['pdict_await_req'])

        
        res_dict = OrderedDict()
        data = self.read_pipe(PIPE_DICT['pdict_await_res'])
        if len(data) != 0 :
            msg = pdict_pb2.PDictVal()
            msg.ParseFromString(data)
            for key, val in msg.val.items():
                res_dict[key] = val.values

        return res_dict
    
    
    def set(self, key, val):
        self.process_pool.apply_async(send_layer_grad_func, (self.did, key, val, PIPE_DICT['pdict_write_req'], PIPE_DICT['pdict_write_res'],))
    
    
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
            fifo = open(fifoname, 'rb')
            self.opened_files[fifoname] = fifo 
        while True: 
            res = fifo.read()
            if len(res) != 0: 
                return res 


    def close(self): 
        for fifo in self.opened_files.values():
            fifo.close()