package production

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Produce struct{}

var p Produce

func InitProduceRouter(router *gin.RouterGroup) {
	produceRouter := router.Group("produce")

	produceRouter.GET("list", p.list)
	produceRouter.GET("fields", p.fields)
	produceRouter.POST("add", p.add)
	produceRouter.POST("update", p.update)
	produceRouter.POST("delete", p.delete)
	produceRouter.POST("void", p.void)
}

func (*Produce) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	produce := &models.Produce{
		OrderNumber: c.DefaultQuery("orderNumber", ""),
		Name:        c.DefaultQuery("name", ""),
		Status:      utils.DefaultQueryInt(c, "status", -1),
	}
	begReportingTime := c.DefaultQuery("begReportingTime", "")
	endReportingTime := c.DefaultQuery("endReportingTime", "")
	begFinishTime := c.DefaultQuery("begFinishTime", "")
	endFinishTime := c.DefaultQuery("endFinishTime", "")

	data, err := service.GetProduceList(produce,
		begReportingTime, endReportingTime,
		begFinishTime, endFinishTime,
		pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Produce) add(c *gin.Context) {
	produce := &models.Produce{}
	if err := c.ShouldBindJSON(produce); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	produce.Operator = c.GetString("userName")
	data, err := service.SaveProduce(produce)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Produce) update(c *gin.Context) {
	produce := &models.Produce{}
	if err := c.ShouldBindJSON(produce); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	produce.Operator = c.GetString("userName")
	data, err := service.UpdateProduce(produce)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Produce) delete(c *gin.Context) {
	produce := &models.Produce{}
	if err := c.ShouldBindJSON(produce); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	produce.Operator = c.GetString("userName")
	err := service.DelProduce(produce.ID, produce.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Produce) void(c *gin.Context) {
	produce := &models.Produce{}
	if err := c.ShouldBindJSON(produce); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	produce.Operator = c.GetString("userName")
	err := service.VoidProduce(produce.ID, produce.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Produce) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetProduceFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
