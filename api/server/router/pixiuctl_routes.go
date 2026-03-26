package router

import (
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/pixiulib/httputils"
)

// downloadPixiuCtlBinary serves pixiuctl binaries from:
// <pixiuctl_binary_dir>/<version>/pixiuctl-<os>-<arch>
func (cr *rainbowRouter) downloadPixiuctl(c *gin.Context) {
	resp := httputils.NewResponse()

	filename := filepath.Base(c.Param("filename"))
	fullPath, err := cr.c.Server().DownloadPixiuctl(c, filepath.Base(c.Param("version")), filename)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(fullPath)
}

func (cr *rainbowRouter) listPixiuctls(c *gin.Context) {
	resp := httputils.NewResponse()

	var err error
	if resp.Result, err = cr.c.Server().ListPixiuctls(c); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	httputils.SetSuccess(c, resp)
}
