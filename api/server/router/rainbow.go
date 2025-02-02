package router

import (
	"github.com/caoyingjunz/rainbow/cmd/app/options"
	"github.com/caoyingjunz/rainbow/pkg/controller"
	"github.com/gin-gonic/gin"
)

type rainbowRouter struct {
	c controller.RainbowInterface
}

// NewRouter initializes a new Task router
func NewRouter(o *options.ServerOptions) {
	s := &rainbowRouter{
		c: o.Controller,
	}
	s.initRoutes(o.HttpEngine)
}

func (cr *rainbowRouter) initRoutes(httpEngine *gin.Engine) {
	taskRoute := httpEngine.Group("/rainbow/tasks")
	{
		taskRoute.POST("", cr.createTask)
		taskRoute.PUT("/:taskId", cr.updateTask)
		taskRoute.DELETE("/:taskId", cr.deleteTask)
		taskRoute.GET("/:taskId", cr.getTask)
		taskRoute.GET("", cr.listTasks)
	}
}
