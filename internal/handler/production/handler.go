package production

import "github.com/gin-gonic/gin"

func InitProductionRouter(router *gin.RouterGroup) {
	productionRouter := router.Group("production")

	InitProduceRouter(productionRouter)
	InitManageRouter(productionRouter)
	InitProduceStockRouter(productionRouter)
}
