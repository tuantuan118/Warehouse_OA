package service

import (
	"github.com/sirupsen/logrus"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func Ticker() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop() // 确保程序退出时停止 ticker

	// 使用循环来处理每次触发的事件
	for range ticker.C {
		// 执行任务
		finished := make([]models.Finished, 0)

		db := global.Db.Model(&models.Finished{})
		db = db.Where("estimated_time <= ?", time.Now())
		db = db.Where("status = ?", 1)
		err := db.Find(&finished).Error
		if err != nil {
			logrus.Infoln("定时任务查找产品库存错误: ", err.Error())
		}

		for _, v := range finished {
			v.Status = 4

			_, err = UpdateFinished(&v)
			if err != nil {
				logrus.Infoln("定时任务修改产品库存错误: ", err.Error())
			}
		}
	}
}
