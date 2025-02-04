package router

import (
	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/gin-gonic/gin"

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

func (cr *rainbowRouter) updateTask(c *gin.Context) {}

func (cr *rainbowRouter) deleteTask(c *gin.Context) {}

func (cr *rainbowRouter) getTask(c *gin.Context) {}

func (cr *rainbowRouter) listTasks(c *gin.Context) {}

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

func (cr *rainbowRouter) updateRegistry(c *gin.Context) {}

func (cr *rainbowRouter) deleteRegistry(c *gin.Context) {}

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

	var err error
	if resp.Result, err = cr.c.Server().ListRegistries(c); err != nil {
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

func (cr *rainbowRouter) listAgents(c *gin.Context) {
	resp := httputils.NewResponse()

	var err error
	if resp.Result, err = cr.c.Server().ListAgents(c); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}
