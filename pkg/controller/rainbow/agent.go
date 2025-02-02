package rainbow

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type AgentGetter interface {
	Agent() Interface
}
type Interface interface {
	Run(ctx context.Context, workers int) error
}

type AgentController struct {
	factory db.ShareDaoFactory
	name    string
	queue   workqueue.RateLimitingInterface
}

func NewAgent(f db.ShareDaoFactory, name string) *AgentController {
	return &AgentController{
		factory: f,
		name:    name,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rainbow-agent"),
	}
}

func (s *AgentController) Run(ctx context.Context, workers int) error {
	// 注册 rainbow 代理
	if err := s.RegisterAgentIfNotExist(ctx); err != nil {
		return err
	}

	go s.getNextWorkItems(ctx)

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 1*time.Second)
	}

	return nil
}

func (s *AgentController) getNextWorkItems(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tasks, err := s.factory.Task().ListWithAgent(ctx, s.name)
		if err != nil {
			klog.Error("failed to list tasks %v", err)
			continue
		}

		if len(tasks) == 0 {
			continue
		}

		for _, task := range tasks {
			s.queue.Add(task.Id)
		}
	}
}

func (s *AgentController) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *AgentController) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	err := s.sync(ctx, key.(int64))
	s.handleErr(ctx, err, key)
	return true
}

func (s *AgentController) sync(ctx context.Context, taskId int64) error {
	task, err := s.factory.Task().Get(ctx, taskId)
	if err != nil {
		return fmt.Errorf("failted to get task %d %v", taskId, err)
	}

	fmt.Println("task:", task)
	return nil
}

// TODO
func (s *AgentController) handleErr(ctx context.Context, err error, key interface{}) {
}

func (s *AgentController) RegisterAgentIfNotExist(ctx context.Context) error {
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
