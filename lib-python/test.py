import torch
import torch.nn as nn
from datetime import datetime

from Dserver.dserver import DServer
import os
import time
# 关于取数据的方式。


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




    

if __name__ == '__main__':
    inv_id = os.environ.get("INV_ID")
    dag_id = os.environ.get("DAG_ID")
    
    model = ResNet50()

    # import ddt 
    # dic = ddt.new(ddt.DICT)
    # dic.set(k, v)
    # dic2 = ddt.import('dic2')
    # dic3 = ddt.merge_sum(dic, dic2)
    # dic3_res = dic3.wait()
    # ddt.dict_set(dic, k, v)

  

    epochs = 1
    batch_size = 32
    train_X = torch.rand([batch_size*100, 3, 32, 32])
    train_Y = torch.randint(1, 100, [batch_size*100])
    train_loader = torch.utils.data.DataLoader(
        torch.utils.data.TensorDataset(train_X, train_Y),
        batch_size=batch_size,
        shuffle=True,
        pin_memory=True)
    # optimizer
    optimizer = torch.optim.Adam(model.parameters(), lr=0.001)

    
    # 注册
    for i in range(epochs):
        train_l_sum = 0
        count = 0
        bp_time = 0
        fp_time = 0
        wait_time = 0
        start_time = datetime.now()
        for X, y in train_loader:
            # 开始注册
            dserver = DServer()
            ddict1 = dserver.new_D("ddict1")
            ddict1.register_params(model.named_parameters())
            ddict2 = dserver.import_D("ddict2")
            ddict3 = dserver.merge_Ds(ddict1, ddict2)
            
            # 注册完成
            print("New Iteration ......")
            fp_start_time = datetime.now()
            optimizer.zero_grad()
            output = model(X)
            loss = torch.nn.CrossEntropyLoss()
            l = loss(output, y)
            train_l_sum += l.cpu().item()
            fp_end_time = datetime.now()
            fp_time += (fp_end_time - fp_start_time).seconds
            
     
            l.backward() # 此时已经将梯度传递出去
            bp_time += (datetime.now() - fp_end_time).seconds
            
            
            # 重置
            ddict1.end() 
            wait_start_time = time.time()
            sum_grads_res = ddict3.wait()
            wait_time += time.time() - wait_start_time
            # 给每层参数换成聚合后的梯度
            for key, val in model.named_parameters():
                val.grad = torch.tensor(sum_grads_res[key]).reshape(val.grad.shape)
            
            # 更新参数
            optimizer.step()
            count += 1
            if count == 1:
                break

        end_time = datetime.now()
        print("avg iter time is: {}", (end_time - start_time).seconds / count)
        print("avg fp iter time is: {}", fp_time / count)
        print("avg bp iter time is: {}", bp_time / count)
        print("avg bp ddict3 wait time is: {}", wait_time / count)
        print('train loss:{}'.format(train_l_sum / count))

    print("end")
