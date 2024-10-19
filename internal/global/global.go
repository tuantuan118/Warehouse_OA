package global

import (
	"gorm.io/gorm"
	"warehouse_oa/internal/config"
)

var (
	Db           *gorm.DB
	ServerConfig = &config.ServerConfig{}
)
