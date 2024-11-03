package order

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Order struct{}

var o Order

func InitOrderRouter(router *gin.RouterGroup) {
	orderRouter := router.Group("order")

	orderRouter.GET("list", o.list)
	orderRouter.GET("fields", o.fields)
	orderRouter.POST("add", o.add)
	orderRouter.POST("update", o.update)
	orderRouter.POST("delete", o.delete)
}

func (*Order) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	order := &models.Order{
		Name:          c.DefaultQuery("name", ""),
		OrderNumber:   c.DefaultQuery("orderNumber", ""),
		Specification: c.DefaultQuery("specification", ""),
		CustomerName:  c.DefaultQuery("customerName", ""),
	}
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetOrderList(order, begTime, endTime, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Order) add(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	data, err := service.SaveOrder(order)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Order) update(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	data, err := service.UpdateOrder(order)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Order) delete(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	err := service.DelOrder(order.ID, order.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Order) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetOrderFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
