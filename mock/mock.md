### 概念
mock 目录下的所有东西都是用于boxwallet下单元测试的模拟数据，只能被 [package]_test 的go文件使用，同时mock中的package依赖关系也体现了项目的启动方式

### 重点
boxwallet/cli目录下有个setup.go文件，如果需要对接不能同时import boxwallet/cli和boxwallet/mock，因为这一个是真实环境启动项依赖，一个是单元测试环境启动依赖项，2个同时起是存在逻辑问题的，这2个package是不能互存的，请大家一定要理解这个设计的概念。
