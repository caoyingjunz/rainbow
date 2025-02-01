package rainbow

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

type AgentGetter interface {
	Agent() Interface
}
type Interface interface {
	Run(ctx context.Context, workers int) error
}

type Agent struct {
	factory db.ShareDaoFactory
	name    string
}

func NewAgent(f db.ShareDaoFactory, name string) *Agent {
	return &Agent{factory: f, name: name}
}

func (s *Agent) Run(ctx context.Context, workers int) error {
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 10*time.Second)
	}

	<-ctx.Done()
	return nil
}

func (s *Agent) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *Agent) processNextWorkItem(ctx context.Context) bool {
	fmt.Println("now", time.Now())
	return true
}
