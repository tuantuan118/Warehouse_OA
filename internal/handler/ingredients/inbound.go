package ingredients

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type InBound struct{}

var ib InBound

func InitInBoundRouter(router *gin.RouterGroup) {
	inBoundRouter := router.Group("in_bound")

	inBoundRouter.GET("list", ib.list)
	inBoundRouter.GET("fields", ib.fields)
	inBoundRouter.POST("add", ib.add)
	inBoundRouter.POST("update", ib.update)
	inBoundRouter.POST("delete", ib.delete)
}

func (*InBound) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	name := c.DefaultQuery("name", "")
	supplier := c.DefaultQuery("supplier", "")
	stockUser := c.DefaultQuery("stockUser", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetInBoundList(name, supplier, stockUser, begTime, endTime, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) add(c *gin.Context) {
	ingredients := &models.IngredientInBound{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.SaveInBound(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) update(c *gin.Context) {
	ingredients := &models.IngredientInBound{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.UpdateInBound(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) delete(c *gin.Context) {
	ingredients := &models.IngredientInBound{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	err := service.DelInBound(ingredients.ID, ingredients.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*InBound) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetInBoundFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}