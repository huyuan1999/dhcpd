package api

import (
	_ "dhcp/api/docs"
	"dhcp/models"
	"dhcp/server"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm/logger"
	"time"
)

var object *models.Object

// @Title DHCP 动态配置 API
// @Description 为 DHCP 服务器提供的简单的 restful api
// @Contact.email 2803660215@qq.com
// @Version 2.2

// @License.name Apache 2.0
// @License.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath
func API(socket string, d *server.DHCPDConfig, logLevel logger.LogLevel, connMaxLifetime time.Duration) {
	object = models.MustConnectDB(d.DBUser, d.DBHost, d.DBPass, d.DBName, d.DBPort, logLevel, d.DBPoolMaxIdleConns, d.DBPoolMaxOpenConns, connMaxLifetime)
	route(socket)
}

func route(socket string) {
	r := gin.Default()
	url := ginSwagger.URL("/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	v1 := r.Group("/api/v1")

	v1.GET("/inform/:tag/", inform)

	v1.POST("/set/options/", setOptions)
	v1.POST("/set/bind/", setBind)
	v1.POST("/set/acl/", setACL)
	v1.POST("/set/reserve/", setReserve)

	v1.PUT("/update/options/", updateOptions)
	v1.PUT("/update/bind/", updateBind)
	v1.PUT("/update/acl/", updateACL)

	v1.DELETE("/del/bind/", deleteBind)
	v1.DELETE("/del/acl/", deleteACL)
	v1.DELETE("/del/reserve/", deleteReserve)

	if err := r.Run(socket); err != nil {
		panic(err)
	}
}
