package product

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Product struct{}

var c Product

func InitProductRouter(router *gin.RouterGroup) {
	productRouter := router.Group("product")

	productRouter.GET("list", c.list)
	productRouter.GET("fields", c.fields)
	productRouter.POST("add", c.add)
	productRouter.POST("update", c.update)
	productRouter.POST("delete", c.delete)
}

func (*Product) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	product := &models.Product{
		OrderNumber: c.DefaultQuery("orderNumber", ""),
		Name:        c.DefaultQuery("name", ""),
	}
	data, err := service.GetProductList(product, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Product) add(c *gin.Context) {
	product := &models.Product{}
	if err := c.ShouldBindJSON(product); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	product.Operator = c.GetString("userName")
	data, err := service.SaveProduct(product)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Product) update(c *gin.Context) {
	product := &models.Product{}
	if err := c.ShouldBindJSON(product); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	product.Operator = c.GetString("userName")
	data, err := service.UpdateProduct(product)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Product) delete(c *gin.Context) {
	product := &models.Product{}
	if err := c.ShouldBindJSON(product); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	product.Operator = c.GetString("userName")
	err := service.DelProduct(product.ID, product.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Product) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetProductFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
