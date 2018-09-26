* tx 中有 input 组，(destination, amount) 组成的 output 组，还有一条找零。
* input 数量 = 支付的数量 + 找零 + 矿工费

### 针对 txPool 的交易

1. 从 tx 中解出对应的 utxo 信息，并保存。其中 blockHeight 为全 1；


### 针对 block 的交易

1. 从 block/txs 中解出对应的 input 和交易信息；
2. 查找库中是否已存在对应的 utxo，若存在，则更新相关数据（blockHeight 等）；
3. 如果不存在，则插入新的记录，并删除已经消耗掉的 input 数据；

### UTXO 的处理流程

1. 以 tx 的 inputs 查找库中已存在的 utxo（sender 对应钱包地址）
2. tx 中 utxo 的信息包含