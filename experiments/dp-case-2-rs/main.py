# import torch
# import torch.nn as nn
# import time
# from Dserver.rs_dserver import DServer
# import os


# # 关于取数据的方式。


# class BasicBlock(nn.Module):
#     """Basic Block for resnet 18 and resnet 34
#     """
#     expansion = 1

#     def __init__(self, in_channels, out_channels, stride=1):
#         super(BasicBlock, self).__init__()

#         self.residual_branch = nn.Sequential(
#             nn.Conv2d(in_channels,
#                       out_channels,
#                       kernel_size=3,
#                       stride=stride,
#                       padding=1,
#                       bias=False), nn.BatchNorm2d(out_channels),
#             nn.ReLU(inplace=True),
#             nn.Conv2d(out_channels,
#                       out_channels * BasicBlock.expansion,
#                       kernel_size=3,
#                       padding=1,
#                       bias=False),
#             nn.BatchNorm2d(out_channels * BasicBlock.expansion))

#         self.shortcut = nn.Sequential()

#         if stride != 1 or in_channels != BasicBlock.expansion * out_channels:
#             self.shortcut = nn.Sequential(
#                 nn.Conv2d(in_channels,
#                           out_channels * BasicBlock.expansion,
#                           kernel_size=1,
#                           stride=stride,
#                           bias=False),
#                 nn.BatchNorm2d(out_channels * BasicBlock.expansion))

#     def forward(self, x):
#         return nn.ReLU(inplace=True)(self.residual_branch(x) +
#                                      self.shortcut(x))


# class BottleNeck(nn.Module):
#     """Residual block for resnet over 50 layers
#     """
#     expansion = 4

#     def __init__(self, in_channels, out_channels, stride=1):
#         super(BottleNeck, self).__init__()
#         self.residual_branch = nn.Sequential(
#             nn.Conv2d(in_channels, out_channels, kernel_size=1, bias=False),
#             nn.BatchNorm2d(out_channels),
#             nn.ReLU(inplace=True),
#             nn.Conv2d(out_channels,
#                       out_channels,
#                       stride=stride,
#                       kernel_size=3,
#                       padding=1,
#                       bias=False),
#             nn.BatchNorm2d(out_channels),
#             nn.ReLU(inplace=True),
#             nn.Conv2d(out_channels,
#                       out_channels * BottleNeck.expansion,
#                       kernel_size=1,
#                       bias=False),
#             nn.BatchNorm2d(out_channels * BottleNeck.expansion),
#         )

#         self.shortcut = nn.Sequential()

#         if stride != 1 or in_channels != out_channels * BottleNeck.expansion:
#             self.shortcut = nn.Sequential(
#                 nn.Conv2d(in_channels,
#                           out_channels * BottleNeck.expansion,
#                           stride=stride,
#                           kernel_size=1,
#                           bias=False),
#                 nn.BatchNorm2d(out_channels * BottleNeck.expansion))

#     def forward(self, x):
#         return nn.ReLU(inplace=True)(self.residual_branch(x) +
#                                      self.shortcut(x))


# class ResNet(nn.Module):
#     def __init__(self, block, layers, num_classes=100, inter_layer=False):
#         super(ResNet, self).__init__()
#         self.inter_layer = inter_layer
#         self.in_channels = 64

#         self.conv1 = nn.Sequential(
#             nn.Conv2d(3, 64, kernel_size=3, padding=1, bias=False),
#             nn.BatchNorm2d(64), nn.ReLU(inplace=True))

#         self.stage2 = self._make_layer(block, 64, layers[0], 1)
#         self.stage3 = self._make_layer(block, 128, layers[1], 2)
#         self.stage4 = self._make_layer(block, 256, layers[2], 2)
#         self.stage5 = self._make_layer(block, 512, layers[3], 2)
#         self.avg_pool = nn.AdaptiveAvgPool2d((1, 1))
#         self.flatten = nn.Flatten()
#         self.fc = nn.Linear(512 * block.expansion, num_classes)

#     def _make_layer(self, block, out_channels, num_blocks, stride):
#         """make resnet layers(by layer i didnt mean this 'layer' was the 
#         same as a neuron netowork layer, ex. conv layer), one layer may 
#         contain more than one residual block 
#         Args:
#             block: block type, basic block or bottle neck block
#             out_channels: output depth channel number of this layer
#             num_blocks: how many blocks per layer
#             stride: the stride of the first block of this layer

#         Return:
#             return a resnet layer
#         """

#         strides = [stride] + [1] * (num_blocks - 1)
#         layers = []
#         for stride in strides:
#             layers.append(block(self.in_channels, out_channels, stride))
#             self.in_channels = out_channels * block.expansion

#         return nn.Sequential(*layers)

#     def forward(self, x):
#         x = self.conv1(x)

#         if self.inter_layer:
#             x1 = self.stage2(x)
#             x2 = self.stage3(x1)
#             x3 = self.stage4(x2)
#             x4 = self.stage5(x3)
#             x = self.avg_pool(x4)
#             x = x.view(x.size(0), -1)
#             x = self.fc(x)

#             return [x1, x2, x3, x4, x]
#         else:
#             x = self.stage2(x)
#             x = self.stage3(x)
#             x = self.stage4(x)
#             x = self.stage5(x)
#             x = self.avg_pool(x)
#             x = self.flatten(x)
#             x = self.fc(x)

#             return x


# def ResNet50():
#     return ResNet(BottleNeck, [3, 4, 6, 3])


# def ResNet101():
#     return ResNet(BottleNeck, [3, 4, 23, 3])


# def ResNet152():
#     return ResNet(BottleNeck, [3, 8, 36, 3])




    

# if __name__ == '__main__': 
#     # model = ResNet50()
#     model = ResNet152()
#     # model = ResNet101()

#     epochs = 1
#     batch_size = 32
#     train_X = torch.rand([batch_size*100, 3, 32, 32])
#     train_Y = torch.randint(1, 100, [batch_size*100])
#     train_loader = torch.utils.data.DataLoader(
#         torch.utils.data.TensorDataset(train_X, train_Y),
#         batch_size=batch_size,
#         shuffle=True,
#         pin_memory=True)
#     # optimizer
#     optimizer = torch.optim.SGD(model.parameters(), lr=0.001)
#     loss = torch.nn.CrossEntropyLoss()

#     dserver = DServer()
    
#     # 注册
#     for i in range(epochs):
#         train_l_sum = 0
#         count = 0
#         iters_time = 0
#         bp_time = 0
#         fp_time = 0
#         load_time = 0
#         ddict_res_wait_time = 0
#         hooks_time = 0
        
#         for X, y in train_loader:
#             # 开始注册
#             start_time = time.time()
#             ddict2 = dserver.new_D("ddict2_"+ str(count))
#             ddict2.register_params(model.named_parameters())
#             dserver.export_D(ddict2, "ddict2_"+ str(count))
#             ddict_res = dserver.import_D("ddict_res_"+ str(count), True)
#             optimizer.zero_grad()
#             # 注册完成
#             print("New Iteration ......", flush=True)
            
#             fp_start_time = time.time()
#             output = model(X)
#             fp_end_time = time.time()
            
#             l = loss(output, y)
#             train_l_sum += l.cpu().item()
#             l.backward() # 此时已经将梯度传递出去
#             bp_end_time = time.time()
            
#             # 重置
#             ddict2.end()  
#             wait_start_time = time.time()
#             sum_grads_res = ddict_res.wait()
#             wait_end_time = time.time()
#             print("Wait successfully", flush=True)
            
#             load_start_time = time.time()
#             for key, val in model.named_parameters():
#                 val.grad = sum_grads_res[key]
#             load_end_time = time.time()
            
#             optimizer.step()
#             end_time = time.time()
            
#             if count>= 1:
#                 fp_time += (fp_end_time - fp_start_time)
#                 bp_time += (bp_end_time - fp_end_time)
#                 ddict_res_wait_time += (wait_end_time - wait_start_time)
#                 load_time += (load_end_time - load_start_time)
#                 iters_time += (end_time - start_time)
#                 hooks_time += ddict2.hook_time
            
#             count += 1
#             if count == 3:
#                 break
        
#         print("avg iter time is: {}".format(iters_time/(count-1)), flush=True)
#         print("avg fp time is: {}".format(fp_time/(count-1)), flush=True)
#         print("avg bp time is: {}".format(bp_time/(count-1)), flush=True)
#         print("avg merged_dict_wait time is: {}".format(ddict_res_wait_time/(count-1)), flush=True)
#         print("avg load time is: {}".format(load_time/(count-1)), flush=True)
#         print("avg hooks time is: {}".format(hooks_time/(count-1)), flush=True)
#         print('train loss:{}'.format(train_l_sum / count), flush=True)
#         # print("avg iter time is: {}".format(iters_time/(count)), flush=True)
#         # print("avg fp time is: {}".format(fp_time/(count)), flush=True)
#         # print("avg bp time is: {}".format(bp_time/(count)), flush=True)
#         # print("avg merged_dict_wait time is: {}".format(ddict_res_wait_time/(count)), flush=True)
#         # print("avg load time is: {}".format(load_time/(count)), flush=True)
#         # print('train loss:{}'.format(train_l_sum / count), flush=True)

#     dserver.close()
#     print("end", flush=True)






import torch
import pickle
import redis
import torch.nn as nn
import time
import threading
import queue

class BasicBlock(nn.Module):
    """Basic Block for resnet 18 and resnet 34
    """
    expansion = 1

    def __init__(self, in_channels, out_channels, stride=1):
        super(BasicBlock, self).__init__()

        self.residual_branch = nn.Sequential(
            nn.Conv2d(in_channels,
                      out_channels,
                      kernel_size=3,
                      stride=stride,
                      padding=1,
                      bias=False), nn.BatchNorm2d(out_channels),
            nn.ReLU(inplace=True),
            nn.Conv2d(out_channels,
                      out_channels * BasicBlock.expansion,
                      kernel_size=3,
                      padding=1,
                      bias=False),
            nn.BatchNorm2d(out_channels * BasicBlock.expansion))

        self.shortcut = nn.Sequential()

        if stride != 1 or in_channels != BasicBlock.expansion * out_channels:
            self.shortcut = nn.Sequential(
                nn.Conv2d(in_channels,
                          out_channels * BasicBlock.expansion,
                          kernel_size=1,
                          stride=stride,
                          bias=False),
                nn.BatchNorm2d(out_channels * BasicBlock.expansion))

    def forward(self, x):
        return nn.ReLU(inplace=True)(self.residual_branch(x) +
                                     self.shortcut(x))




class BottleNeck(nn.Module):
    """Residual block for resnet over 50 layers
    """
    expansion = 4

    def __init__(self, in_channels, out_channels, stride=1):
        super(BottleNeck, self).__init__()
        self.residual_branch = nn.Sequential(
            nn.Conv2d(in_channels, out_channels, kernel_size=1, bias=False),
            nn.BatchNorm2d(out_channels),
            nn.ReLU(inplace=True),
            nn.Conv2d(out_channels,
                      out_channels,
                      stride=stride,
                      kernel_size=3,
                      padding=1,
                      bias=False),
            nn.BatchNorm2d(out_channels),
            nn.ReLU(inplace=True),
            nn.Conv2d(out_channels,
                      out_channels * BottleNeck.expansion,
                      kernel_size=1,
                      bias=False),
            nn.BatchNorm2d(out_channels * BottleNeck.expansion),
        )

        self.shortcut = nn.Sequential()

        if stride != 1 or in_channels != out_channels * BottleNeck.expansion:
            self.shortcut = nn.Sequential(
                nn.Conv2d(in_channels,
                          out_channels * BottleNeck.expansion,
                          stride=stride,
                          kernel_size=1,
                          bias=False),
                nn.BatchNorm2d(out_channels * BottleNeck.expansion))

    def forward(self, x):
        return nn.ReLU(inplace=True)(self.residual_branch(x) +
                                     self.shortcut(x))


class ResNet(nn.Module):
    def __init__(self, block, layers, num_classes=100, inter_layer=False):
        super(ResNet, self).__init__()
        self.inter_layer = inter_layer
        self.in_channels = 64

        self.conv1 = nn.Sequential(
            nn.Conv2d(3, 64, kernel_size=3, padding=1, bias=False),
            nn.BatchNorm2d(64), nn.ReLU(inplace=True))

        self.stage2 = self._make_layer(block, 64, layers[0], 1)
        self.stage3 = self._make_layer(block, 128, layers[1], 2)
        self.stage4 = self._make_layer(block, 256, layers[2], 2)
        self.stage5 = self._make_layer(block, 512, layers[3], 2)
        self.avg_pool = nn.AdaptiveAvgPool2d((1, 1))
        self.flatten = nn.Flatten()
        self.fc = nn.Linear(512 * block.expansion, num_classes)

    def _make_layer(self, block, out_channels, num_blocks, stride):
        """make resnet layers(by layer i didnt mean this 'layer' was the 
        same as a neuron netowork layer, ex. conv layer), one layer may 
        contain more than one residual block 
        Args:
            block: block type, basic block or bottle neck block
            out_channels: output depth channel number of this layer
            num_blocks: how many blocks per layer
            stride: the stride of the first block of this layer
        
        Return:
            return a resnet layer
        """

        strides = [stride] + [1] * (num_blocks - 1)
        layers = []
        for stride in strides:
            layers.append(block(self.in_channels, out_channels, stride))
            self.in_channels = out_channels * block.expansion

        return nn.Sequential(*layers)

    def forward(self, x):
        x = self.conv1(x)

        if self.inter_layer:
            x1 = self.stage2(x)
            x2 = self.stage3(x1)
            x3 = self.stage4(x2)
            x4 = self.stage5(x3)
            x = self.avg_pool(x4)
            x = x.view(x.size(0), -1)
            x = self.fc(x)

            return [x1, x2, x3, x4, x]
        else:
            x = self.stage2(x)
            x = self.stage3(x)
            x = self.stage4(x)
            x = self.stage5(x)
            x = self.avg_pool(x)
            x = self.flatten(x)
            x = self.fc(x)

            return x


def ResNet50():
    return ResNet(BottleNeck, [3, 4, 6, 3])

def ResNet101():
    return ResNet(BottleNeck, [3, 4, 23, 3])

def ResNet152():
    return ResNet(BottleNeck, [3, 8, 36, 3])

import struct
def write_pb_message(pb_message):
    serialized_message = pb_message.SerializeToString()
    size = len(serialized_message)
    res = struct.pack('<I', size) + serialized_message
    return res

def read_pb_message_from_file(bytes_msg, start, pb_message):
    size_data = bytes_msg[start:start+4]
    if not size_data:
        return 0
    start += 4
    size = struct.unpack('<I', size_data)[0]
    serialized_message = bytes_msg[start:start+size]
    pb_message.ParseFromString(serialized_message)
    start += size
    return start

MODEL_MAP = {
    "Resnet50": ResNet50(),
    "Resnet101": ResNet101(),
    "Resnet152": ResNet152()
} 

def write_thread(q):
    while True:
        data = q.get() 
        msg = pickle.dumps(data) 
        # fifo.write(msg)
        # fifo.flush()

class ModelWorker(object):
    def __init__(self, total_partitions: int, self_partition_id: int, model_name : str, version: str, epochs: int, partition_plan: dict):
        '''
          partition_plan = {
              0: [0,1,2]
              1: [3]
              2: [4,5,6,7]
          }
        '''
        self.model_name = model_name
        self.partition_id = self_partition_id
        self.total_partitions = total_partitions
        self.version = version
        self.model = MODEL_MAP[model_name]
        self.epochs = epochs
        self.hook_time = 0 
        
        self.write_q = queue.Queue(300)
        import os 
        fifoname = "hello.fifo"
        if not os.path.exists(fifoname):
            os.mkfifo(fifoname)  
        # self.write_fifo = open(fifoname, "wb") 
        self.writer = threading.Thread(target=write_thread, args=(self.write_q, ))
        self.writer.start()
        
        self.loss_fn = torch.nn.CrossEntropyLoss()
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=0.001)
        self.train_data = None
        self.train_data_loader = None

        self.rconn = redis.Redis(host='redis.openfaas', port=6379, health_check_interval=30) 
        
        
        self.dump_time = 0
        self.load_time = 0
        self.reshape_encode_list_time = 0
        self.reshape_decode_tensor_time = 0
        self.upload_time = 0
        
        self.tensor_shape_dict = {}
        for key, val in self.model.named_parameters():
            self.tensor_shape_dict[key] = val.shape 
            
        # 构建dataloaders
        self.init_data_loader()

            
    def init_data_loader(self):
        batch_size = 32
        train_X = torch.rand([batch_size*100, 3, 32, 32])
        train_Y = torch.randint(1,100,[batch_size*100])
        self.train_data_loader = torch.utils.data.DataLoader(
            torch.utils.data.TensorDataset(train_X, train_Y),
            batch_size=batch_size,
            shuffle=True,
            pin_memory=True)
    
    
    def train(self):
        # 前向反向以及下载-合并-上传
        res_str = ""
        
        # 注册hook_function
        def hook_factory(name):
            def hook(grad):
                hook_start_time = time.time()
                # self.write_q.put((name, grad))
                # pickle.dumps((name, grad))
                hook_end_time = time.time()
                self.hook_time += (hook_end_time - hook_start_time)
            return hook
        
        # for name, param in self.model.named_parameters():
        #     hook = param.register_hook(hook_factory(name))
      
            
        
        for epoch in range(self.epochs):
            train_l_sum = 0
            iter_count = 0
            start_time = time.time()
            wait_time = 0
            fp_time = 0
            bp_time = 0
            merged_time = 0
            download_time = 0
            iters_time = 0
            hooks_time = 0

            for X, y in self.train_data_loader:
                iter_start_time = time.time()
                self.optimizer.zero_grad()
                fp_start_time = time.time()
                output = self.model(X)
                fp_end_time = time.time()
                

                l = self.loss_fn(output, y)
                train_l_sum += l.cpu().item()
                l.backward()
                bp_end_time = time.time()
                
                
                iter_count += 1
                print("iter: {}, fp time:{} , bp time: {}".format(iter_count, fp_end_time - fp_start_time, bp_end_time - fp_end_time))

                # 如果不是，将全部的梯度上传，等待合并后的结果；
                # 如果是，不上传，等待所有的worker的梯度，全部下载下来，并将合并后的结果上传Redis；
                if self.partition_id == 0: #master
                    # wait for all uploaded
                    left_download_set = set([i for i in range(self.total_partitions) if i!= self.partition_id])
                    grads_list = []
                    wait_start_time = time.time()
                    while(True):
                        tmp_set = set()
                        for i in left_download_set: 
                            download_key = "{}-{}-{}-{}".format(self.version, epoch, iter_count, i)
                            download_start = time.time()
                            download_res = self.rconn.get(name=download_key) 
                            download_time += (time.time() - download_start)
                            if download_res != None : 
                                grads = self.load_vals(download_res)
                                grads_list.append(grads)
                                tmp_set.add(i)
                        left_download_set = left_download_set - tmp_set
                        if len(left_download_set) == 0:
                            break
                        # time.sleep(0.1)
                    wait_end_time = time.time()
                    # wait_time += (wait_end_time-wait_start_time)
                    print("downloaded all parts. ")
                    
                    # merge
                    merge_start_time = time.time()
                    merged_grads = {}
                    for key, param in self.model.named_parameters():
                        key_sum = 0
                        for i in range(len(grads_list)):
                            key_sum = key_sum + grads_list[i][key]
                        key_sum = key_sum + param.grad
                        avg_grad = key_sum/self.total_partitions
                        param.grad = avg_grad
                        merged_grads[key] = avg_grad
                    merge_end_time = time.time()
                    merged_time += (merge_end_time - merge_start_time)
                    
                    
                    # upload merged
                    upload_key_prefix = "{}-{}-{}-merged".format(self.version, epoch, iter_count)
                    self.upload_all_partitions(upload_key_prefix) 
                    print("uploaded merged. ")
                    
                    # delete last iteration result 
                    delete_keys = [i for i in range(self.total_partitions) if i!= self.partition_id]
                    for i in delete_keys:
                        delete_key =  "{}-{}-{}-{}".format(self.version, epoch, iter_count, i)
                        self.rconn.delete(delete_key)
                    if iter_count>=2:
                        self.rconn.delete("{}-{}-{}-merged".format(self.version, epoch, iter_count-1))
                    # apply merged grads
                    self.optimizer.step()
                    
                else:
                    # upload
                    upload_key_prefix = "{}-{}-{}-{}".format(self.version, epoch, iter_count, self.partition_id)
                    self.upload_all_partitions(upload_key_prefix) 
                    print("uploaded. key_prefix: {}".format(upload_key_prefix))
                    
                    # wait for merged
                    wait_start_time = time.time()
                    merged_grads = None
                    download_key = "{}-{}-{}-merged".format(self.version, epoch, iter_count)
                    while(True):
                        download_start = time.time()
                        download_res = self.rconn.get(name=download_key) 
                        download_time += (time.time() - download_start)
                        if download_res != None : 
                            merged_grads = self.load_vals(download_res)
                            break
                        # time.sleep(0.1)
                    wait_end_time = time.time()
                    # wait_time += (wait_end_time-wait_start_time)
                    
                    # apply merged grads                       
                    for key, val in self.model.named_parameters():
                        val.grad = merged_grads[key]
                    self.optimizer.step()
                    
                # 更新完成
                iter_end_time = time.time()
                # if iter_count == 1:
                #     self.rconn.delete("{}-{}-{}-merged".format(self.version, epoch, iter_count))
                #     break # 就来一个iteration
                if iter_count >= 2:
                    # add time 
                    fp_time += (fp_end_time - fp_start_time)
                    bp_time += (bp_end_time - fp_end_time)
                    wait_time += (wait_end_time - wait_start_time)
                    iters_time += (iter_end_time - iter_start_time)
                    
                    pass
                
                if iter_count == 5:
                    self.rconn.delete("{}-{}-{}-merged".format(self.version, epoch, iter_count))
                    break
                    
                
                
            # print('epoch: {}, train_loss: {:.4f}'.format(epoch, train_l_sum / iter_count))

            res_str = "avg iter time is {}".format(iters_time/(iter_count-1), flush=True)
            print("avg iter time is: {}".format(iters_time/(iter_count-1)), flush=True)
            print("avg fp time is: {}".format(fp_time/(iter_count-1)), flush=True)
            print("avg bp time is: {}".format(bp_time/(iter_count-1)), flush=True)
            print("avg wait time is: {}".format(wait_time/(iter_count-1)), flush=True)
            print("avg hook time is: {}".format(self.hook_time/(iter_count)), flush=True)
            
        return res_str

    

    
    # def upload_all_partitions(self, key_prefix):
    #     dump_start_time = time.time()
        
    #     val = bytearray()
    #     for name, param in  self.model.named_parameters():
    #         new_param = param.reshape(-1).tolist()
    #         msg = layer_msg_pb2.layer_msg()
    #         msg.layer_name = name
    #         msg.tensor_len = len(new_param)
    #         msg.values.extend(new_param)
    #         val.extend(write_pb_message(msg))
    #     val = bytes(val)
    #     dump_end_time = time.time()
    #     self.dump_time += (dump_end_time - dump_start_time)
        
        
    #     # exit(0)
    #     upload_start = time.time()
    #     self.rconn.set(key_prefix, val)
    #     self.upload_time += (time.time() - upload_start)
    
    def upload_all_partitions(self, key_prefix):
        dump_start_time = time.time()
        grad_dict = {}
        for name, param in  self.model.named_parameters():
            grad_dict[name] = param.grad
        data = pickle.dumps(grad_dict)
        dump_end_time = time.time()
        self.dump_time += (dump_end_time - dump_start_time)
        
        upload_start = time.time()
        self.rconn.set(key_prefix, data)
        self.upload_time += (time.time() - upload_start)


    # def load_vals(self, pb_encoded):
    #     load_start_time = time.time()
    #     res = OrderedDict()
        
    #     index = 0
    #     while True:
    #         layer_msg = layer_msg_pb2.layer_msg()

    #         size_data = pb_encoded[index:index+4]
    #         if not size_data:
    #             break
    #         index += 4
    #         size = struct.unpack('<I', size_data)[0]
    #         serialized_message = pb_encoded[index : index+size]
    #         layer_msg.ParseFromString(serialized_message)
    #         index += size
    #         # layer_msgs.append(layer_msg)
    #         key = layer_msg.layer_name
    #         val = layer_msg.values
    #         res[key] = torch.tensor(val).reshape(self.tensor_shape_dict[key])
        
    #     self.load_time += (time.time()-load_start_time)
         
    #     return res

    def load_vals(self, pkl_data):
        load_start_time = time.time()
        grad_dict = pickle.loads(pkl_data)
        self.load_time += (time.time()-load_start_time)
        return grad_dict







plan_dict ={
    0:[0,1,2,3],
    1:[4,5,6,7]
}

partition_id = 1
total_partitions = 2
worker = ModelWorker(total_partitions, partition_id, "Resnet101", "20230526-2workers", 1, plan_dict)
res_str = worker.train()




# model = ResNet101()

            
    
# batch_size = 32
# train_X = torch.rand([batch_size*100, 3, 32, 32])
# train_Y = torch.randint(1,100,[batch_size*100])
# train_data_loader = torch.utils.data.DataLoader(
#     torch.utils.data.TensorDataset(train_X, train_Y),
#     batch_size=batch_size,
#     shuffle=True,
#     pin_memory=True
# )
# loss_fn = torch.nn.CrossEntropyLoss()
# optimizer = torch.optim.SGD(model.parameters(), lr=0.001)

# count = 0
# for X, y in train_data_loader:
#         iter_start_time = time.time()
#         optimizer.zero_grad()
#         fp_start_time = time.time()
#         output = model(X)
#         fp_end_time = time.time()
        
#         l = loss_fn(output, y)
#         l.backward()
#         bp_end_time = time.time() 
#         count += 1 
#         print("iter {}: fp_time is {}, bp_time is {}".format(count, fp_end_time-fp_start_time, bp_end_time-fp_end_time))
        
#         if count == 10:
#             break
   

           
                
                