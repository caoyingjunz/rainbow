package rainbow

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"k8s.io/klog/v2"
)

// ScoreInterface
type ScoreInterface interface {
	Name() string
	Weight() int64
	Score(ctx context.Context, agent *model.Agent, task *model.Task) (int64, error)
}

// Scheduler 调度器，负责选择最优的 Agent
type Scheduler struct {
	plugins []ScoreInterface
}

func NewScheduler(plugins ...ScoreInterface) *Scheduler {
	return &Scheduler{
		plugins: plugins,
	}
}

func (s *Scheduler) AddPlugin(plugin ScoreInterface) {
	s.plugins = append(s.plugins, plugin)
}

// Select best agent
func (s *Scheduler) SelectBestAgent(ctx context.Context, agents []*model.Agent, task *model.Task) (*model.Agent, error) {
	if len(agents) == 0 {
		return nil, fmt.Errorf("no available agents")
	}

	// 为每个 agent 计算总分
	type agentScore struct {
		agent *model.Agent
		score int64
	}

	var agentScores []agentScore

	for _, agent := range agents {
		totalScore := int64(0)

		// 计算每个插件的分数
		for _, plugin := range s.plugins {
			score, err := plugin.Score(ctx, agent, task)
			if err != nil {
				klog.Warningf("Plugin %s failed to score agent %s: %v", plugin.Name(), agent.Name, err)
				continue
			}

			// 应用权重
			weightedScore := score * plugin.Weight()
			totalScore += weightedScore

			klog.V(4).Infof("Agent %s scored %d by plugin %s (weight: %d)",
				agent.Name, score, plugin.Name(), plugin.Weight())
		}

		agentScores = append(agentScores, agentScore{
			agent: agent,
			score: totalScore,
		})

		klog.V(3).Infof("Agent %s total score: %d", agent.Name, totalScore)
	}

	// 按分数排序，选择分数最高的
	sort.Slice(agentScores, func(i, j int) bool {
		return agentScores[i].score > agentScores[j].score
	})

	bestAgent := agentScores[0].agent
	klog.Infof("Selected agent %s with score %d", bestAgent.Name, agentScores[0].score)

	return bestAgent, nil
}

// AvailabilityScore 可用性打分插件
type AvailabilityScore struct{}

func (a *AvailabilityScore) Name() string {
	return "availability plugin"
}

func (a *AvailabilityScore) Weight() int64 {
	return 100
}

func (a *AvailabilityScore) Score(ctx context.Context, agent *model.Agent, task *model.Task) (int64, error) {
	// 检查 agent 状态
	switch agent.Status {
	case model.RunAgentType:
		return 100, nil
	case model.UnknownAgentType:
		return 0, nil
	case model.UnRunAgentType:
		return 0, nil
	default:
		return 50, nil
	}
}

type LoadBalanceScore struct {
	factory db.ShareDaoFactory
}

func NewLoadBalanceScore(factory db.ShareDaoFactory) *LoadBalanceScore {
	return &LoadBalanceScore{
		factory: factory,
	}
}

func (l *LoadBalanceScore) Name() string {
	return "load_balance plugin"
}

func (l *LoadBalanceScore) Weight() int64 {
	return 80
}

func (l *LoadBalanceScore) Score(ctx context.Context, agent *model.Agent, task *model.Task) (int64, error) {
	// 获取该 agent 当前运行的任务数
	runningTasks, err := l.factory.Task().GetRunningTask(ctx, db.WithName(agent.Name))
	if err != nil {
		return 0, err
	}

	// 统计该 agent 的任务数
	agentTaskCount := len(runningTasks)

	// 计算负载分数：任务数越少，分数越高
	maxConcurrency := 10
	if agentTaskCount >= maxConcurrency {
		return 0, nil
	}

	// 使用负指数函数计算分数，任务数越多分数越低
	score := int64(100 * math.Exp(-float64(agentTaskCount)/2.0))
	return score, nil
}

type ResourceScore struct {
	factory db.ShareDaoFactory
}

func NewResourceScore(factory db.ShareDaoFactory) *ResourceScore {
	return &ResourceScore{
		factory: factory,
	}
}

func (r *ResourceScore) Name() string {
	return "resource"
}

func (r *ResourceScore) Weight() int64 {
	return 60
}

func (r *ResourceScore) Score(ctx context.Context, agent *model.Agent, task *model.Task) (int64, error) {
	// 检查 agent 的 GitHub 账号开销
	if agent.GrossAmount >= 16.0 {
		// 达到上限，分数很低
		return 10, nil
	}

	// 根据开销金额计算分数
	remainingBudget := 16.0 - agent.GrossAmount
	score := int64(remainingBudget * 6.25) // 16 * 6.25 = 100

	if score > 100 {
		score = 100
	}

	return score, nil
}

type TaskTypeScore struct{}

func (t *TaskTypeScore) Name() string {
	return "task_type plugin"
}

func (t *TaskTypeScore) Weight() int64 {
	return 30
}

func (t *TaskTypeScore) Score(ctx context.Context, agent *model.Agent, task *model.Task) (int64, error) {
	// 根据任务类型和 agent 类型进行匹配
	if task.IsPublic {
		// 公开镜像任务优先分配给 public 类型的 agent
		if agent.Type == model.PublicAgentType {
			return 100, nil
		} else {
			return 60, nil
		}
	} else {
		// 私有镜像任务优先分配给 private 类型的 agent
		if agent.Type == model.PrivateAgentType {
			return 100, nil
		} else {
			return 60, nil
		}
	}
}
