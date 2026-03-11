package rainbow

import (
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) SearchRepo(ctx *gin.Context, listOption types.ListOptions) (interface{}, error) {
	path, tag, err := ParseImageItem(listOption.NameSelector)
	if err != nil {
		return nil, err
	}
	arch := listOption.Arch
	if len(arch) == 0 {
		arch = defaultArch
	}

	// 优先从自己的仓库查询，如果没有，则从官方仓库查询
	tags, err := s.factory.Image().SearchTags(ctx, tag, arch, path, listOption.UserId)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("record not found")
	}

	return tags[0], nil
}
