package user

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
)

func InitUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("user")

	userRouter.GET("ping", ping)
	userRouter.POST("login", login)
	userRouter.POST("register", register)
}

func login(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	data, err := service.Login(user.Name, user.Password)
	if err != nil {
		handler.InternalServerError(c, err)
	}

	handler.Success(c, data)
}

func register(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	data, err := service.Register(user)
	if err != nil {
		handler.InternalServerError(c, err)
	}

	handler.Success(c, data)
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
