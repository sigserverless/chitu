from collections import OrderedDict
import os
import errno
import json
from Dserver.constants import PIPE_DICT

class DDict:
    
    def __init__(self, did, opened_files) -> None:
        self.data_dict = OrderedDict()
        self.did = did
        self.total_layers = 0
        self.curr_layer_count = 0
        self.hooks = []
        self.opened_files = opened_files

    
    def register_params(self, params_iterator):
        self.managed_param_dict = {key : val for key, val in params_iterator}
        self.total_layers = len(self.managed_param_dict) 
        self.gen_idx_to_key_dict()
        def hook_func(grad):
            self.set(self.idx_to_layer_name[self.curr_layer_count], grad.reshape(-1).tolist())
            self.curr_layer_count += 1
         
        
        for _, val in self.managed_param_dict.items():
            hook = val.register_hook(hook_func)
            self.hooks.append(hook)
        
    
    def end(self):
        handle_msg = json.dumps({
            "Id" : self.did  # 还需要其它的字段吗
        })
        self.write_pipe(handle_msg, PIPE_DICT['end_req'])
        self.reset()

    
    def reset(self):
        self.curr_layer_count = 0
        self.data_dict = OrderedDict()
        for hook in self.hooks:
            hook.remove()

    
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
        self.write_pipe(handle_msg, PIPE_DICT['await_req'])

        res_dict = OrderedDict()
        
        data = self.read_pipe(PIPE_DICT['await_res'])
        if len(data) != 0 :
            data = json.loads(data)['Val']['Val']
            for key, val in data.items():
                res_dict[key] = val

        return res_dict
    
    
    def set(self, key, val):
        self.data_dict[key] = val
        data = json.dumps({
            "Id":  self.did ,
            "Wop": {
                "kind" : "DictSet",
                "val" : {
                    "K" : key,
                    "V" : val,
                }
            }            
        })
        self.write_pipe(data, PIPE_DICT['write_req'])
        # Next line shall be removed to save a bit time. But the buffered 
        # responses in the fifo must be read somehow. So next line might 
        # be replaced by a asynchronous call or a continuous thread which 
        # repeats to read the fifo.
        self.read_pipe(PIPE_DICT['write_res'])
        print("successfully set key: {}".format(key), flush=True)
   
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


    def close(self): 
        for fifo in self.opened_files.values():
            fifo.close()