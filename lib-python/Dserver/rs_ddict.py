# from collections import OrderedDict
# import time
# import threading
# import pickle

# class DDict:
    
#     def __init__(self, dname = "", dag_id = "" , write_q = None, read_q = None) -> None:
#         self.data_dict = dict()
#         self.set_counts_dict = dict()
#         self.write_q = write_q
#         self.read_q = read_q
#         self.dname = dname
#         self.dag_id = dag_id
#         self.completed_keys_count = 0
#         self.hook_time = 0
#         self.total_keys = -1
#         self.hooks = []
#         self.observers = []
#         self.watchees = []
#         self.EXPORT = False
#         self.IMPORT = False
#         self.MERGED = False
#         self.END = False
#         self.key_lock = threading.Lock()

#     def register_watchee(self, watchee):
#         self.watchees.append(watchee)

#     def register_observer(self, observer):
#         self.observers.append(observer)
        
#     def onchange(self, key, val):
#         # merge logic
#         with self.key_lock:
            
#             if key not in self.data_dict:
#                 self.data_dict[key] = val
#                 self.set_counts_dict[key] = 1
#             else :
#                 # print("XXXXonchange key: {}, key_counts: {}".format(key, self.set_counts_dict[key]), flush=True)
#                 self.data_dict[key] = (self.data_dict[key]*self.set_counts_dict[key] + val) / (self.set_counts_dict[key]+1)
#                 self.set_counts_dict[key] += 1
#                 # print("YYYYonchange key: {}, key_counts: {}".format(key, self.set_counts_dict[key]), flush=True)
#             # send "pub" to other workers
#             if self.set_counts_dict[key] == len(self.watchees):
#                 self.completed_keys_count += 1
#                 if self.EXPORT: # if worker-1's merged_ddict, broadcast
#                     bytes_grad = pickle.dumps([key, self.data_dict[key]])
#                     bytes_len = len(bytes_grad)
#                     data = {
#                         "kind": "pub",
#                         "key": self.dname,
#                         "len": bytes_len
#                     }
#                     # print("merged key-{} into write_q... ".format(key), flush=True)
#                     self.write_q.put(data)
#                     self.write_q.put(bytes_grad)
#                 if self.completed_keys_count == self.total_keys:
#                     data = {
#                         "kind": "pub",
#                         "key": self.dname,
#                         "len": 0
#                     }
#                     self.write_q.put(data)
#                     self.END = True
        
#     def change(self, key, val): # change for remote dict
#         for observer in self.observers:
#             observer.onchange(key, val)
    
#     def read_fifo(self):
#         self.read_thread = threading.Thread(target=self.read_fifo_thread)
#         self.read_thread.start()
        
#     def read_fifo_thread(self): 
#         while True:
#             message = self.read_q.get()
#             if message[0] != "end":
                
#                 if not self.MERGED: # if ddict2 from worker2
#                     # self.change(message[0], message[1])
#                     # for transfer one time
#                     for name, grad in message.items():
#                         self.change(name, grad)
                    
#                 else :  # if merged_dict from worker1
#                     # self.data_dict[message[0]] = message[1]
#                     self.data_dict = message
#             else: 
#                self.END = True 
#                break
    

#     def register_params(self, params_iterator):
#         def hook_factory(name):
#             def hook(grad):
#                 hook_start = time.time()
#                 if self.EXPORT:
#                     # if ddict2 of worker2, broadcast
#                     # write_pipe
#                     self.data_dict[name] = True
#                     bytes_grad = pickle.dumps([name, grad])
#                     bytes_len = len(bytes_grad)
#                     data = {
#                         "kind": "pub",
#                         "key": self.dname,
#                         "len": bytes_len
#                     }
#                     # print("worker2 layer is computed: ", name)
#                     self.write_q.put(data)
#                     self.write_q.put(bytes_grad)
                    
#                     if len(self.data_dict) == self.total_keys:
#                         data = {
#                         "kind": "pub",
#                         "key": self.dname,
#                         "len": 0
#                         }
#                         self.write_q.put(data)
#                         self.END = True
                    
#                 else:
#                     # if ddict1 of worker1, onchange only 
#                     self.change(name, grad)
#                     # print("worker1 layer is computed : ", name)
#                 hook_end = time.time()
#                 self.hook_time += (hook_end - hook_start)
#             return hook
        
#         count = 0
#         for name, param in params_iterator:
#             # hook = param.register_hook(hook_factory(name))
#             # self.hooks.append(hook)
#             self.data_dict[name] = param
#             count+=1
#         self.total_keys = count    
    
#     def end(self):
#         # self.data_dict = OrderedDict()
        
#         # update 
#         if self.EXPORT:
#             # if ddict2 of worker2, broadcast
#             # write_pipe
#             res_dict = {}
#             for name, param in self.data_dict.items():
#                 res_dict[name] = param.grad
#             bytes_grad = pickle.dumps(res_dict)
#             bytes_len = len(bytes_grad)
#             data = {
#                 "kind": "pub",
#                 "key": self.dname,
#                 "len": bytes_len
#             }
#             self.write_q.put(data)
#             self.write_q.put(bytes_grad)
            
#             # write end
#             data = {
#             "kind": "pub",
#             "key": self.dname,
#             "len": 0
#             }
#             self.write_q.put(data)
#             self.END = True
            
#         else:
#             # if ddict1 of worker1, onchange only 
#             for name, param in self.data_dict.items():
#                 self.change(name, param.grad)
        
#         for hook in self.hooks:
#             hook.remove()

#     def wait(self):
#         while True:
#             time.sleep(0.001)
#             if self.END == True: 
#                 break
#         return self.data_dict
            
    
 
#     def close(self): 
#         for fifo in self.opened_files.values():
#             fifo.close()





from collections import OrderedDict
import time
import threading
import pickle

class DDict:
    
    def __init__(self, dname = "", dag_id = "" , write_q = None, read_q = None) -> None:
        self.data_dict = dict()
        self.set_counts_dict = dict()
        self.write_q = write_q
        self.read_q = read_q
        self.dname = dname
        self.dag_id = dag_id
        self.completed_keys_count = 0
        self.hook_time = 0
        self.total_keys = -1
        self.hooks = []
        self.observers = []
        self.watchees = []
        self.EXPORT = False
        self.IMPORT = False
        self.MERGED = False
        self.END = False
        self.key_lock = threading.Lock()
        self.transfer_dict = dict()

    def register_watchee(self, watchee):
        self.watchees.append(watchee)

    def register_observer(self, observer):
        self.observers.append(observer)
        
    def onchange(self, key, val):
        # merge logic
        with self.key_lock:
            
            if key not in self.data_dict:
                self.data_dict[key] = val
                self.set_counts_dict[key] = 1
            else :
                self.data_dict[key] = (self.data_dict[key]*self.set_counts_dict[key] + val) / (self.set_counts_dict[key]+1)
                self.set_counts_dict[key] += 1
            # send "pub" to other workers
            if self.set_counts_dict[key] == len(self.watchees):
                self.completed_keys_count += 1
                self.transfer_dict[key] = self.data_dict[key]
                if self.completed_keys_count% 100 == 0 or self.completed_keys_count == self.total_keys:
                    bytes_grad = pickle.dumps(self.transfer_dict)
                    bytes_len = len(bytes_grad)
                    data = {
                        "kind": "pub",
                        "key": self.dname,
                        "len": bytes_len
                    }
                    self.write_q.put(data)
                    self.write_q.put(bytes_grad)
                    # self.write_q.put((self.dname, self.transfer_dict))
                    self.transfer_dict = dict()
                     
                if self.completed_keys_count == self.total_keys:
                    # send 'end' signal 
                    data = {
                        "kind": "pub",
                        "key": self.dname,
                        "len": 0
                    }
                    self.write_q.put(data)
                    self.END = True
        
    def change(self, key, val): # change for remote dict
        for observer in self.observers:
            observer.onchange(key, val)
    
    def read_fifo(self):
        self.read_thread = threading.Thread(target=self.read_fifo_thread)
        self.read_thread.start()
        
    def read_fifo_thread(self): 
        while True:
            message = self.read_q.get()
            if type(message) is dict: 
                if not self.MERGED: # if ddict2 from worker2
                    # self.change(message[0], message[1])
                    # for transfer one time
                    for name, grad in message.items():
                        self.change(name, grad)
                    
                else :  # if merged_dict from worker1
                    # self.data_dict[message[0]] = message[1]
                    # self.data_dict = message
                    self.data_dict.update(message)
            else: 
               self.END = True 
               break
    

    def register_params(self, params_iterator):
        def hook_factory(name):
            def hook(grad):
                hook_start = time.time()
                if self.EXPORT:
                    # if ddict2 of worker2, broadcast
                    # write_pipe
                    self.data_dict[name] = grad
                    if len(self.data_dict) == 100:
                        bytes_grad = pickle.dumps(self.data_dict)
                        bytes_len = len(bytes_grad)
                        data = {
                            "kind": "pub",
                            "key": self.dname,
                            "len": bytes_len
                        }
                        self.write_q.put(data)
                        self.write_q.put(bytes_grad)
                        self.data_dict = {} #置空
                else:
                    # if ddict1 of worker1, onchange only 
                    self.change(name, grad)
                hook_end = time.time()
                self.hook_time += (hook_end - hook_start)
            return hook
        
        count = 0
        for name, param in params_iterator:
            hook = param.register_hook(hook_factory(name))
            self.hooks.append(hook)
            count+=1
        self.total_keys = count    
    
    def end(self):
        
        # update 
        if self.EXPORT:
            # if ddict2 of worker2, broadcast
            # write_pipe
            if len(self.data_dict) != 0:
                bytes_grad = pickle.dumps(self.data_dict)
            bytes_len = len(bytes_grad)
            data = {
                "kind": "pub",
                "key": self.dname,
                "len": bytes_len
            }
            self.write_q.put(data)
            self.write_q.put(bytes_grad)
            
            # write end
            data = {
                "kind": "pub",
                "key": self.dname,
                "len": 0
            }
            self.write_q.put(data)
            self.END = True
            
        else:
            # if ddict1 of worker1, onchange only 
            for name, param in self.data_dict.items():
                self.change(name, param.grad)
        
        for hook in self.hooks:
            hook.remove()

    def wait(self):
        while True:
            time.sleep(0.001)
            if self.END == True: 
                break
        return self.data_dict
            
    
 
    def close(self): 
        for fifo in self.opened_files.values():
            fifo.close()