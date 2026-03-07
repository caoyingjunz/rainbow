package rainbow

import (
	"context"
	"fmt"

	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) SearchRepo(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	targetRepo := listOption.NameSelector

	fmt.Println("targetRepo", targetRepo)

	fmt.Println("arch", listOption.Arch)
	return nil, nil
}
