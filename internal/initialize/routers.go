package initialize

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler/customer"
	"warehouse_oa/internal/handler/ingredients"
	"warehouse_oa/internal/handler/order"
	"warehouse_oa/internal/handler/production"
	"warehouse_oa/internal/handler/user"
	"warehouse_oa/internal/middlewares"
)

func InitRouters() *gin.Engine {
	Router := gin.Default()
	Router.Use(middlewares.Cors())

	apiGroup := Router.Group("/api/v1")
	user.InitLoginRouter(apiGroup)

	group := apiGroup
	group.Use(middlewares.JWTAuth())
	{
		user.InitUserRouter(group)
		user.InitRoleRouter(group)

		customer.InitCustomerRouter(group)
		ingredients.InitIngredientRouter(group)
		production.InitProductionRouter(group)
		order.InitOrderRouter(group)
	}

	user.InitPermissionRouter(group)
	return Router
}
