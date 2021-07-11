package main

import (
	"dhcp/api"
	"dhcp/models"
	"dhcp/server"
	"flag"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"time"
)

var (
	d         *server.DHCPDConfig
	enableApi bool
	apiListen string
)

func init() {
	d = &server.DHCPDConfig{}

	flag.BoolVar(&enableApi, "enable-api", true, "打开 api 接口")
	flag.StringVar(&apiListen, "api-listen", "0.0.0.0:8888", "api 接口监听地址")

	flag.StringVar(&d.Listen, "dhcpd-listen", "0.0.0.0", "dhcpd 监听地址")
	flag.IntVar(&d.Port, "dhcpd-port", 67, "dhcpd 监听端口")
	flag.StringVar(&d.IFName, "dhcpd-ifname", "", "dhcpd 监听接口")
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

func setLogLevel() logger.LogLevel {
	log.SetReportCaller(true)
	log.SetLevel(log.WarnLevel)
	dbLogLevel := logger.Error
	if d.Debug {
		dbLogLevel = logger.Info
		log.SetLevel(log.DebugLevel)
	}
	return dbLogLevel
}

// 每分钟检查一次租约表, 删除过期的租约信息
func DeleteExpiredLease(object *models.Object) {
	c := cron.New()
	_, err := c.AddFunc("* * * * *", func() {
		options := server.QueryOptions()
		leaseTime, err := time.ParseDuration(options.LeaseTime)
		if err != nil {
			log.Fatalf("Error delete expired lease lease generation time %s", err.Error())
			return
		}
		object.Db.Unscoped().Where("unix_timestamp(expires) < ?", time.Now().Add(leaseTime).Unix()).Delete(&models.Leases{})
	})
	if err != nil {
		log.Fatalf("Error init delete expired lease cron job %s", err.Error())
	}
	c.Start()
}

// 设置 dhcp 服务器默认默认参数(仅在 Options 表为空的时候调用)
func createDefaultConfig(object *models.Object) {
	options := models.Options{
		LeaseTime:    "1h",
		ServerIP:     "10.1.1.1",
		BootFileName: "pxelinux.0",
		GatewayIP:    "10.1.1.1",
		RangeStartIP: "10.1.1.10",
		RangeEndIP:   "10.1.1.100",
		NetMask:      "255.0.0.0",
		Router:       "10.1.1.1",
		DNS:          "223.5.5.5,223.6.6.6",
		ACL:          false,
		ACLAction:    "",
	}
	object.Db.FirstOrCreate(&options)
}

func main() {
	logLevel := setLogLevel()
	connMaxLifetime := time.Second * time.Duration(d.DBPoolConnMaxLifetime)

	object := models.MustConnectDB(d.DBUser, d.DBHost, d.DBPass, d.DBName, d.DBPort, logLevel, d.DBPoolMaxIdleConns, d.DBPoolMaxOpenConns, connMaxLifetime)

	if err := object.Db.AutoMigrate(&models.Leases{}, &models.Options{}, &models.ACL{}, &models.Binding{}, &models.Reserves{}); err != nil {
		panic(err)
	}

	createDefaultConfig(object)

	go DeleteExpiredLease(object)

	go func() {
		if enableApi {
			api.API(apiListen, d, logLevel, connMaxLifetime)
		}
	}()

	// 启动 dhcpd
	server.DHCPD(d, logLevel, connMaxLifetime)
}
