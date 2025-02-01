package rainbow

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
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
	// 注册 rainbow 代理
	if err := s.RegisterAgentIfNotExist(ctx); err != nil {
		return err
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

func (s *Agent) RegisterAgentIfNotExist(ctx context.Context) error {
	if len(s.name) == 0 {
		return fmt.Errorf("agent name missing")
	}

	var err error
	_, err = s.factory.Agent().GetByName(ctx, s.name)
	if err == nil {
		return nil
	}
	_, err = s.factory.Agent().Create(ctx, &model.Agent{Name: s.name})
	return err
}
