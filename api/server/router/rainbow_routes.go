package router

import (
	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (cr *rainbowRouter) createTask(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req types.CreateTaskRequest
		err error
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().CreateTask(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) updateTask(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req    types.UpdateTaskRequest
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, &req, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	req.Id = idMeta.ID
	if err = cr.c.Server().UpdateTask(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) UpdateTaskStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		req    types.UpdateTaskStatusRequest
		err    error
	)
	if err = httputils.ShouldBindAny(c, &req, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	req.TaskId = idMeta.ID
	if err = cr.c.Server().UpdateTaskStatus(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) deleteTask(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().DeleteTask(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) getTask(c *gin.Context) {}

func (cr *rainbowRouter) listTasks(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		userMeta types.UserMeta
		err      error
	)
	if err = httputils.ShouldBindAny(c, nil, nil, &userMeta); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().ListTasks(c, userMeta.UserId); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) createRegistry(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req types.CreateRegistryRequest
		err error
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if err = cr.c.Server().CreateRegistry(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) updateRegistry(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		req    types.UpdateRegistryRequest
		err    error
	)
	if err = httputils.ShouldBindAny(c, &req, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	req.Id = idMeta.ID
	if err = cr.c.Server().UpdateRegistry(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) deleteRegistry(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().DeleteRegistry(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) getRegistry(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().GetRegistry(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) listRegistries(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err      error
		userMeta types.UserMeta
	)
	if err = httputils.ShouldBindAny(c, nil, nil, &userMeta); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().ListRegistries(c, userMeta.UserId); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) getAgent(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().GetAgent(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) updateAgentStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req types.UpdateAgentStatusRequest
		err error
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().UpdateAgentStatus(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) listAgents(c *gin.Context) {
	resp := httputils.NewResponse()

	var err error
	if resp.Result, err = cr.c.Server().ListAgents(c); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) createImage(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req types.CreateImageRequest
		err error
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().CreateImage(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) updateImage(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		req    types.UpdateImageRequest
		err    error
	)
	if err = httputils.ShouldBindAny(c, &req, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	req.Id = idMeta.ID
	if err = cr.c.Server().UpdateImage(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) deleteImage(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().SoftDeleteImage(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) UpdateImageStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req types.UpdateImageStatusRequest
		err error
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().UpdateImageStatus(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) getImage(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().GetImage(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) listImages(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		userMeta types.UserMeta
		taskMeta types.TaskMeta
		err      error
	)
	if err = httputils.ShouldBindAny(c, nil, nil, &userMeta); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().ListImages(c, taskMeta.TaskId, userMeta.UserId); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}
