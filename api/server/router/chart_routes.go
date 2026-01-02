package router

import (
	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/gin-gonic/gin"
)

func (cr *rainbowRouter) enableChartRepo(c *gin.Context) {
	resp := httputils.NewResponse()
	var (
		err error
		req types.EnableChartRepoRequest
	)
	if err = httputils.ShouldBindAny(c, &req, nil, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if err = cr.c.Server().EnableChartRepo(c, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) ListCharts(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err        error
		req        types.ChartMetaRequest
		listOption types.ListOptions
	)
	if err = httputils.ShouldBindAny(c, nil, &req, &listOption); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	listOption.Project = req.Project
	listOption.SetDefaultPageOption()
	if resp.Result, err = cr.c.Server().ListCharts(c, listOption); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) DeleteChart(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err error
		req types.ChartMetaRequest
	)
	if err = httputils.ShouldBindAny(c, nil, &req, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if err = cr.c.Server().DeleteChart(c, req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) ListChartVersions(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err        error
		req        types.ChartMetaRequest
		listOption types.ListOptions
	)
	if err = httputils.ShouldBindAny(c, nil, &req, &listOption); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	listOption.SetDefaultPageOption()
	if resp.Result, err = cr.c.Server().ListChartTags(c, req, listOption); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) GetChartVersion(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err error
		req types.ChartMetaRequest
	)
	if err = httputils.ShouldBindAny(c, nil, &req, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if resp.Result, err = cr.c.Server().GetChartTag(c, req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) DeleteChartVersion(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err error
		req types.ChartMetaRequest
	)
	if err = httputils.ShouldBindAny(c, nil, &req, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if err = cr.c.Server().DeleteChartTag(c, req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}

func (cr *rainbowRouter) uploadChart(c *gin.Context) {
	resp := httputils.NewResponse()

	var (
		err error
		req types.ChartMetaRequest
	)
	if err = httputils.ShouldBindAny(c, nil, &req, nil); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	if err = cr.c.Server().UploadChart(c, req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	httputils.SetSuccess(c, resp)
}
