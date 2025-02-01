package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

type RainbowAgentGetter interface {
	RainbowAgent() Interface
}
type Interface interface {
	Run(ctx context.Context, workers int) error
}

type RainbowAgent struct {
	factory db.ShareDaoFactory
}

func NewRainbowAgent(f db.ShareDaoFactory) *RainbowAgent {
	return &RainbowAgent{factory: f}
}

func (s *RainbowAgent) Run(ctx context.Context, workers int) error {
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, time.Second)
	}

	<-ctx.Done()
	return nil
}

func (s *RainbowAgent) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *RainbowAgent) processNextWorkItem(ctx context.Context) bool {

	return true
}
