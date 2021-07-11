这是一个专门为pxe装机实现的dhcp服务器，它可能不满足任何的标准

部分代码来源：https://github.com/coredhcp/coredhcp

#### 功能
* dhcp 核心功能（包括分配IP，主机名，路由，网关，DNS等）
* 基于restful api的动态配置
* mac 地址绑定
* 对 pxe 的支持
* acl 黑白名单
* 所有状态数据存储在数据库中（当前只支持mysql）
* 基于 restful api 的动态配置
* 基于 swagger 的 api 文档


#### 部署
* 操作系统: CentOS 7.5
* 数据库: MySQL 5.7
* go 1.16
```bash
# 编译
$ go build

# 查看帮助信息
$ ./dhcp --help

# 启动 dhcpd 和 api
$ ./dhcp --db-pass=xxx --dhcpd-ifname=em1

# 打开 swagger 文档
http://127.0.0.1:8888/swagger/index.html
```
