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
	rainbowdRoute := httpEngine.Group("/rainbow/rainbowds")
	{
		rainbowdRoute.GET("", cr.listRainbowds)
	}

	taskRoute := httpEngine.Group("/rainbow/tasks")
	{
		taskRoute.POST("", cr.createTask)
		taskRoute.PUT("/:Id", cr.updateTask)
		taskRoute.DELETE("/:Id", cr.deleteTask)
		taskRoute.GET("/:Id", cr.getTask)
		taskRoute.GET("", cr.listTasks)

		taskRoute.PUT("/:Id/status", cr.UpdateTaskStatus) // DEPRECATED
		taskRoute.GET(":Id/images", cr.listTaskImages)
		taskRoute.POST("/rerun", cr.reRunTask)

		taskRoute.POST("/:Id/messages", cr.createTaskMessage)
		taskRoute.GET(":Id/messages", cr.listTaskMessages)
	}

	archRoute := httpEngine.Group("/rainbow/architectures")
	{
		archRoute.GET("", cr.listArchitectures)
	}

	subscribeRoute := httpEngine.Group("/rainbow/subscribes")
	{
		subscribeRoute.POST("", cr.createSubscribe)
		subscribeRoute.PUT("/:Id", cr.updateSubscribe)
		subscribeRoute.DELETE("/:Id", cr.deleteSubscribe)
		subscribeRoute.GET("/:Id", cr.getSubscribe)
		subscribeRoute.GET("", cr.listSubscribes)
		subscribeRoute.GET("/:Id/messages", cr.listSubscribeMessages)
		subscribeRoute.POST("/:Id/run", cr.runSubscribeImmediately)
	}

	kubernetesVersionRoute := httpEngine.Group("/rainbow/kubernetes/versions")
	{
		kubernetesVersionRoute.GET("", cr.listKubernetesVersions)
		// 废弃
		kubernetesVersionRoute.POST("/sync", cr.syncRemoteKubernetesVersions)
	}

	registryRoute := httpEngine.Group("/rainbow/registries")
	{
		registryRoute.POST("", cr.createRegistry)
		registryRoute.POST("/login", cr.loginRegistry)
		registryRoute.PUT("/:Id", cr.updateRegistry)
		registryRoute.DELETE("/:Id", cr.deleteRegistry)
		registryRoute.GET("/:Id", cr.getRegistry)
		registryRoute.GET("", cr.listRegistries)
	}

	agentRoute := httpEngine.Group("/rainbow/agents")
	{
		agentRoute.POST("", cr.createAgent)
		agentRoute.PUT("/:Name", cr.updateAgent)
		agentRoute.DELETE("/:Id", cr.deleteAgent)
		agentRoute.GET("/:Id", cr.getAgent)
		agentRoute.GET("", cr.listAgents)
		agentRoute.PUT("/status", cr.updateAgentStatus)

		// 指定 agent 创建 repo
		agentRoute.POST("/:Name/repos", cr.createAgentRepo)
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
		imageRoute.DELETE("/:Id/tags/:TagId", cr.deleteImageTag)
	}

	batchRoute := httpEngine.Group("/rainbow/batch")
	{
		batchRoute.POST("/list/images", cr.listImagesByIds)     // 根据ids批量获取镜像列表
		batchRoute.POST("/delete/images", cr.deleteImagesByIds) // 根据ids批量删除镜像列表

		batchRoute.POST("/list/tasks", cr.listTasksByIds)
		batchRoute.POST("/delete/tasks", cr.deleteTasksByIds)
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

	userRoute := httpEngine.Group("/rainbow/users")
	{
		userRoute.POST("", cr.createUser)
		userRoute.PUT("/:Id", cr.updateUser)
		userRoute.DELETE("/:Id", cr.deleteUser)
		userRoute.GET("/:Id", cr.getUser)
		userRoute.GET("", cr.listUsers)
	}
	syncRoute := httpEngine.Group("/rainbow/sync")
	{
		syncRoute.POST("/users", cr.createOrUpdateUsers)
		syncRoute.POST("/kubernetes/versions", cr.syncRemoteKubernetesVersions)
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
		repoRoute.GET("/repositories/tags", cr.searchRepositoryTags)
		repoRoute.GET("/repositories/:namespace/:name/tags/:tag", cr.searchRepositoryTagInfo)
	}

	notifyRoute := httpEngine.Group("/rainbow/notifications")
	{
		notifyRoute.POST("", cr.createNotification)
		notifyRoute.PUT("/:Id", cr.updateNotification)
		notifyRoute.DELETE("/:Id", cr.deleteNotification)
		notifyRoute.GET("/:Id", cr.getNotification)
		notifyRoute.GET("", cr.listNotifications)

		notifyRoute.POST("/enable/notify", cr.enableNotify)
	}

	notifyTypeRoute := httpEngine.Group("/rainbow/notification/types")
	{
		notifyTypeRoute.GET("", cr.getNotificationTypes)
	}

	sendNotifyRoute := httpEngine.Group("/rainbow/send/notification")
	{
		sendNotifyRoute.POST("", cr.sendNotification)
	}

	fixRoute := httpEngine.Group("/rainbow/fix")
	{
		fixRoute.POST("", cr.fix)
	}

	chartRoute := httpEngine.Group("/rainbow/chartrepo")
	{
		chartRoute.POST("/enable", cr.enableChartRepo) // 启用 helm chart repo
		chartRoute.GET("/:project/status", cr.getChartRepoStatus)

		chartRoute.GET("/:project/charts", cr.ListCharts)
		chartRoute.GET("/:project/charts/:chart", cr.ListChartVersions)
		chartRoute.DELETE("/:project/charts/:chart", cr.DeleteChart)

		chartRoute.GET("/:project/token", cr.getToken)

		// 上传 chart 到指定项目
		chartRoute.POST("/upload/:project", cr.uploadChart)
		// 下载 chart
		chartRoute.GET("/download/:project/charts/:chart/:version", cr.downloadChart)

		chartRoute.GET("/:project/charts/:chart/:version", cr.GetChartVersion)
		chartRoute.DELETE("/:project/charts/:chart/:version", cr.DeleteChartVersion)
	}

	// 构建镜像 API
	buildRoute := httpEngine.Group("/rainbow/builds")
	{
		buildRoute.POST("", cr.createBuild)
		buildRoute.DELETE("/:Id", cr.deleteBuild)
		buildRoute.PUT("/:Id", cr.updateBuild)
		buildRoute.GET("", cr.listBuilds)
		buildRoute.GET("/:Id", cr.getBuild)

		buildRoute.POST("/:Id/messages", cr.createBuildMessage)
		buildRoute.GET("/:Id/messages", cr.listBuildMessages)
	}

	// 设置资源状态API
	setStatus := httpEngine.Group("/rainbow/set")
	{
		setStatus.PUT("/build/:Id/status", cr.setBuildStatus)
		setStatus.PUT("/task/:Id/status", cr.UpdateTaskStatus)
	}
}
