package rainbowd

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type RainbowdGetter interface {
	Rainbowd() Interface
}

type Interface interface {
	Run(ctx context.Context, workers int) error
}

type rainbowdController struct {
	name    string
	factory db.ShareDaoFactory
	cfg     rainbowconfig.Config

	queue workqueue.RateLimitingInterface
}

func New(f db.ShareDaoFactory, cfg rainbowconfig.Config) *rainbowdController {
	return &rainbowdController{
		factory: f,
		cfg:     cfg,
		name:    cfg.Agent.Name,
	}
}

func (s *rainbowdController) Run(ctx context.Context, workers int) error {
	if err := s.RegisterIfNotExist(ctx); err != nil {
		klog.Errorf("register rainbowd failed: %v", err)
		return err
	}

	go s.getNextWorkItems(ctx)

}

func (s *rainbowdController) RegisterIfNotExist(ctx context.Context) error {
	if len(s.name) == 0 {
		return fmt.Errorf("rainbowd name is empty")
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 1*time.Second)
	}

	return nil
}

func (s *rainbowdController) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *rainbowdController) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	taskId, resourceVersion, err := KeyFunc(key)
	klog.Infof("任务(%v)被调度到本节点，即将开始处理", key)
	if err != nil {
		s.handleErr(ctx, err, key)
	} else {
		_ = s.factory.Task().UpdateDirectly(ctx, taskId, map[string]interface{}{"status": "镜像初始化", "message": "初始化环境中", "process": 1})
		if err = s.factory.Task().CreateTaskMessage(ctx, &model.TaskMessage{TaskId: taskId, Message: "节点调度完成"}); err != nil {
			klog.Errorf("记录节点调度失败 %v", err)
		}
		if err = s.sync(ctx, taskId, resourceVersion); err != nil {
			if msgErr := s.factory.Task().CreateTaskMessage(ctx, &model.TaskMessage{TaskId: taskId, Message: fmt.Sprintf("同步失败，原因: %v", err)}); msgErr != nil {
				klog.Errorf("记录同步失败 %v", msgErr)
			}
			s.handleErr(ctx, err, key)
		}
	}
	return true
}

func (s *rainbowdController) getNextWorkItems(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取未处理
		tasks, err := s.factory.Task().ListWithAgent(ctx, s.name, 0)
		if err != nil {
			klog.Error("failed to list tasks %v", err)
			continue
		}
		if len(tasks) == 0 {
			continue
		}

		for _, task := range tasks {
			s.queue.Add(fmt.Sprintf("%d/%d", task.Id, task.ResourceVersion))
		}
	}
}

func KeyFunc(key interface{}) (int64, int64, error) {
	str, ok := key.(string)
	if !ok {
		return 0, 0, fmt.Errorf("failed to convert %v to string", key)
	}
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("parts length not 2")
	}

	taskId, err := strutil.ParseInt64(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to Parse taskId to Int64 %v", err)
	}
	resourceVersion, err := strutil.ParseInt64(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to Parse resourceVersion to Int64 %v", err)
	}

	return taskId, resourceVersion, nil
}
