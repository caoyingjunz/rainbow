package router

import (
	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (cr *rainbowRouter) createBuild(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req types.CreateBuildRequest
		err error
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().CreateBuild(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) deleteBuild(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err = cr.c.Server().DeleteBuild(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) updateBuild(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		req    types.UpdateBuildRequest
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, &req, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	req.Id = idMeta.ID

	if err = cr.c.Server().UpdateBuild(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) listBuilds(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err        error
		listOption types.ListOptions
	)
	if err = httputils.ShouldBindAny(c, nil, nil, &listOption); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if resp.Result, err = cr.c.Server().ListBuilds(c, listOption); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) getBuild(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		err    error
	)
	if err = httputils.ShouldBindAny(c, nil, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if resp.Result, err = cr.c.Server().GetBuild(c, idMeta.ID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) setBuildStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		idMeta types.IdMeta
		req    types.UpdateBuildStatusRequest
		err    error
	)
	if err = httputils.ShouldBindAny(c, &req, &idMeta, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	req.BuildId = idMeta.ID
	if err = cr.c.Server().UpdateBuildStatus(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}
