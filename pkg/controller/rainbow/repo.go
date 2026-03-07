package rainbow

import (
	"context"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) SearchRepo(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	path, tag, err := ParseImageItem(listOption.NameSelector)
	if err != nil {
		return nil, err
	}
	arch := listOption.Arch
	if len(arch) == 0 {
		arch = defaultArch
	}

	obj, err := s.factory.Image().GetTagBy(ctx, db.WithArchitecture(arch), db.WithPath(path), db.WithName(tag))
	if err != nil {
		return nil, err
	}
	return obj, nil
}
