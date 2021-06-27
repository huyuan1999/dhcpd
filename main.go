package main

import (
	"dhcp/restful"
	"dhcp/server"
	"flag"
)

var (
	d         *server.DHCPDConfig
	enableApi bool
	apiListen string
)

func init() {
	d = &server.DHCPDConfig{}

	flag.StringVar(&d.Listen, "dhcpd-listen", "0.0.0.0", "dhcpd 监听地址")
	flag.IntVar(&d.Port, "dhcpd-port", 67, "dhcpd 监听端口")
	flag.StringVar(&d.IFName, "dhcpd-ifname", "", "dhcpd 监听接口")
	flag.BoolVar(&enableApi, "enable-api", true, "打开 api 接口")
	flag.StringVar(&apiListen, "api-listen", "0.0.0.0:8888", "api 接口监听地址")
	flag.BoolVar(&d.Debug, "debug", false, "是否打开调试日志")

	// init db
	flag.StringVar(&d.DBUser, "db-user", "root", "数据库用户名")
	flag.StringVar(&d.DBHost, "db-host", "127.0.0.1", "数据库地址")
	flag.IntVar(&d.DBPort, "db-port", 3306, "数据库端口")
	flag.StringVar(&d.DBPass, "db-pass", "", "数据库密码")
	flag.StringVar(&d.DBName, "db-name", "dhcpd", "数据库名称")

	// init db connect pool
	flag.IntVar(&d.DBPoolMaxIdleConns, "db-pool-maxidleconns", 10, "连接池中最大空闲连接数")
	flag.IntVar(&d.DBPoolMaxOpenConns, "db-pool-maxopenconns", 100, "打开数据库连接的最大数量")
	flag.IntVar(&d.DBPoolConnMaxLifetime, "db-pool-connmaxlifetime", 3600, "连接可复用的最大时间, 单位秒(s)")
	flag.Parse()
}

func main() {
	control := make(chan struct{}, 1)

	go func() {
		// 等待 dhcpd 前置准备完成(包括数据库初始化, listener 初始化等)
		<- control
		// 启动 restful api
		restful.API(":8888")
	}()

	// 启动 dhcpd
	server.DHCPD(d, control)
}
