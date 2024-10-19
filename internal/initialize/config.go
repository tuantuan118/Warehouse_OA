package initialize

import (
	"warehouse_oa/internal/global"
)

func InitConfig() error {
	global.ServerConfig.MysqlInfo.Host = "127.0.0.1"
	global.ServerConfig.MysqlInfo.Port = 3306
	global.ServerConfig.MysqlInfo.Username = "xxxx"
	global.ServerConfig.MysqlInfo.Password = "xxxxxx"
	global.ServerConfig.MysqlInfo.DbName = "xxxx"

	return nil
}
