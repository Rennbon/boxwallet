### 概念
bccoin是维护各个币种的详细信息的package，目录下有一个json文件，用于实现各币种的换算，整个钱包对于币种计算必须要用这个包，这里控制好了计算精度不丢失，主链币必须通过json文件配置，erc20代币系列通过获取代币信息的时候动态注入，所有的代币信息会存入kv数据库，大致json格式如下：
```
[
  {
    "coinType": 1,    #币种类型，与bccore.BloclChainType对应
    "token": "",      #代币唯一标识符
    "symbol": "BTC",  #币种象征字符串
    "decimals": 8,    #币种精度
    "name": "比特币"   #币种名称
  },
]
```

### 目录介绍
- distribute/coin_info.json   币种信息配置项，这里的配置会在初始化的时候载入
- coin.go  币种宏观的操作，如new一个币种金额，还有CoinAmounter interface下的一些方法
- coin_cache.go  币种数据库操作相关，用户存储币种信息，这里的db接口放在util中，所以其他项目是需要使用这个package可以import bccoin和util就可以了
- init.go 初始化coin_info.json信息到kv数据库用

### 接入新币种如何做到拓展
1. 主链币
> 在./distribute/coin_info.json中添加币种信息就好了，但是要区分好如何添加最合适

2. 代币
> 在bctrans/token目录下完成代币的信息的获取后调用coin_cache.go中的SaveCoinInfo就好了

### 其他项目如何直接复用bccoin
这是一个幸运的拓展项，这里的bccoin是完全解耦的，以下是bccoin的必须启动函数，参数db使用的是util工具包的数据库接口定义，prefix前缀动态传入
```
func InitCoinCache(db util.Database, prefix []byte) *CoinCache {
	if defualtCoinCache == nil {
		defualtCoinCache = &CoinCache{db: db, prefix: prefix}
		defualtCoinCache.BatchSaveCoinInfos(loadCoinInfo())
	}
	return defualtCoinCache
}

```