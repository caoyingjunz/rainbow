package rainbow

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

type RainbowAgent struct {
}

func NewRainbowAgent(ctx context.Context) (*RainbowAgent, error) {
	return &RainbowAgent{}, nil
}

func (s *RainbowAgent) Run(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, time.Second)
	}

	<-ctx.Done()
}

func (s *RainbowAgent) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *RainbowAgent) processNextWorkItem(ctx context.Context) bool {

	return true
}
