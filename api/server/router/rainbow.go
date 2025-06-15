package router

import (
	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
	"github.com/caoyingjunz/rainbow/pkg/controller"
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

		taskRoute.PUT("/:Id/status", cr.UpdateTaskStatus)
		taskRoute.GET(":Id/images", cr.listTaskImages)
		taskRoute.POST("/rerun", cr.reRunTask)

		taskRoute.POST("/:Id/messages", cr.createTaskMessage)
		taskRoute.GET(":Id/messages", cr.listTaskMessages)
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
		agentRoute.PUT("/status", cr.updateAgentStatus)
	}

	imageRoute := httpEngine.Group("/rainbow/images")
	{
		imageRoute.POST("", cr.createImage)
		imageRoute.PUT("/:Id", cr.updateImage)
		imageRoute.DELETE("/:Id", cr.deleteImage)
		imageRoute.GET("/:Id", cr.getImage)
		imageRoute.GET("", cr.listImages)

		imageRoute.PUT("/status", cr.UpdateImageStatus)
		imageRoute.POST("/batches", cr.createImages)
		imageRoute.DELETE("/:Id/tags/:name", cr.deleteImageTag)
	}

	searchRoute := httpEngine.Group("/rainbow/search")
	{
		searchRoute.GET("/public/images", cr.listPublicImages)
	}

	collectRoute := httpEngine.Group("/rainbow/collections")
	{
		collectRoute.GET("", cr.getCollections)
		collectRoute.POST("/add/review", cr.AddDailyReview)
	}

	labelRoute := httpEngine.Group("/rainbow/labels")
	{
		labelRoute.POST("", cr.createLabel)
		labelRoute.DELETE("/:Id", cr.deleteLabel)
		labelRoute.PUT("/:Id", cr.updateLabel)
		labelRoute.GET("", cr.listLabels)
	}

	logoRoute := httpEngine.Group("/rainbow/logos")
	{
		logoRoute.POST("", cr.createLogo)
		logoRoute.DELETE("/:Id", cr.deleteLogo)
		logoRoute.PUT("/:Id", cr.updateLogo)
		logoRoute.GET("", cr.listLogos)
	}

	nsRoute := httpEngine.Group("/rainbow/namespaces")
	{
		nsRoute.POST("", cr.createNamespace)
		nsRoute.PUT("/:Id", cr.updateNamespace)
		nsRoute.DELETE("/:Id", cr.deleteNamespace)
		nsRoute.GET("", cr.listNamespaces)
	}

	// 镜像汇总
	overviewRoute := httpEngine.Group("/rainbow/overview")
	{
		overviewRoute.GET("", cr.overview)
		overviewRoute.GET("/downflow/daily", cr.downflow)
		overviewRoute.GET("/store/daily", cr.store)
		overviewRoute.GET("/image/daily", cr.getImageDownflow)
	}

	repoRoute := httpEngine.Group("/rainbow/search")
	{
		repoRoute.GET("/repositories", cr.searchRepositories)
		repoRoute.GET("/repositories/:namespace/:name/tags", cr.searchRepositoryTags)
	}
}
