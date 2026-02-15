# Controller 模板

> 位置: `internal/controllers/YOUR_entity_controller.go`

## 模板

```go
package controllers

type IYOUR_ENTITYController interface { common.IBaseController }

type yourEntityControllerImpl struct {
	YOUR_ENTITYService services.IYOUR_ENTITYService `inject:""`
	LoggerMgr          loggermgr.ILoggerManager     `inject:""`
}

func NewYOUR_ENTITYController() IYOUR_ENTITYController {
	return &yourEntityControllerImpl{}
}

func (c *yourEntityControllerImpl) ControllerName() string { return "YOUR_ENTITYController" }
func (c *yourEntityControllerImpl) GetRouter() string {
	return "/api/YOUR_TABLE [POST],/api/YOUR_TABLE [GET],/api/YOUR_TABLE/:id [GET],/api/YOUR_TABLE/:id [DELETE]"
}

func (c *yourEntityControllerImpl) Handle(ctx *gin.Context) {
	switch ctx.Request.Method {
	case "POST":
		c.handleCreate(ctx)
	case "GET":
		if ctx.Param("id") != "" { c.handleGet(ctx) } else { c.handleList(ctx) }
	case "DELETE":
		c.handleDelete(ctx)
	}
}

func (c *yourEntityControllerImpl) handleCreate(ctx *gin.Context) {
	var req struct{ Field1, Field2 string }
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	entity, err := c.YOUR_ENTITYService.Create(req.Field1, req.Field2)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": entity})
}

func (c *yourEntityControllerImpl) handleGet(ctx *gin.Context) {
	entity, err := c.YOUR_ENTITYService.GetByID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(404, gin.H{"error": "未找到"})
		return
	}
	ctx.JSON(200, gin.H{"data": entity})
}

func (c *yourEntityControllerImpl) handleList(ctx *gin.Context) {
	entities, err := c.YOUR_ENTITYService.GetAll()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": entities})
}

func (c *yourEntityControllerImpl) handleDelete(ctx *gin.Context) {
	if err := c.YOUR_ENTITYService.Delete(ctx.Param("id")); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"message": "删除成功"})
}
```
