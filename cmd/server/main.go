package main

import (
	"github.com/sirupsen/logrus"
	initialize2 "warehouse_oa/internal/initialize"
)

func main() {
	if err := initialize2.InitConfig(); err != nil {
		logrus.Panicf("init config err:%s", err.Error())
	}
	if err := initialize2.InitDb(); err != nil {
		logrus.Panicf("init db err:%s", err.Error())
	}

	router := initialize2.InitRouters()
	err := router.Run()
	if err != nil {
		logrus.Fatalln("Failed to start router", err.Error())
		return
	} // 监听并在 0.0.0.0:8080 上启动服务
}
