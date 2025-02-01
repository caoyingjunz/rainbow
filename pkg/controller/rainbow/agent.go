package rainbow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/caoyingjunz/rainbow/pkg/db"
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
	if len(s.name) == 0 {
		return fmt.Errorf("agent name missing")
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 2*time.Second)
	}

	return nil
}

func (s *Agent) worker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.processNextWorkItem()
	}
}

func (s *Agent) processNextWorkItem() {
	fmt.Println("now", time.Now())
}
