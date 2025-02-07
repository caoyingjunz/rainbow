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
		taskRoute.PUT("/:Id", cr.updateTask)
		taskRoute.DELETE("/:Id", cr.deleteTask)
		taskRoute.GET("/:Id", cr.getTask)
		taskRoute.GET("", cr.listTasks)
	}

	registryRoute := httpEngine.Group("/rainbow/registries")
	{
		registryRoute.POST("", cr.createRegistry)
		registryRoute.PUT("/:Id", cr.updateRegistry)
		registryRoute.DELETE("/:Id", cr.deleteRegistry)
		registryRoute.GET("/:Id", cr.getRegistry)
		registryRoute.GET("", cr.listRegistries)
	}

	agentRoute := httpEngine.Group("/rainbow/agents")
	{
		agentRoute.GET("/:Id", cr.getAgent)
		agentRoute.GET("", cr.listAgents)
	}

	imageRoute := httpEngine.Group("/rainbow/images")
	{
		imageRoute.POST("", cr.createImage)
		imageRoute.PUT("", cr.updateImage)
		imageRoute.GET("/:Id", cr.getImage)
		imageRoute.GET("", cr.listImages)
	}
}
