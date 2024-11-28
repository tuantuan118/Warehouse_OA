package finished

import "github.com/gin-gonic/gin"

func InitProductionRouter(router *gin.RouterGroup) {
	productionRouter := router.Group("finished")

	InitFinishedRouter(productionRouter)
	InitManageRouter(productionRouter)
	InitFinishedStockRouter(productionRouter)
}
