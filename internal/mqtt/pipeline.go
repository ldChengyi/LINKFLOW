package mqtt

import (
	"context"

	"github.com/ldchengyi/linkflow/internal/model"
)

// Node 管道节点接口
type Node interface {
	Name() string
	Process(ctx context.Context, pc *PipelineContext) error
}

// PipelineContext 管道流转上下文
type PipelineContext struct {
	// 输入
	RawText  string
	DeviceID string // 发送指令的设备ID
	UserID   string

	// 预处理
	Tokens []string

	// 意图分类
	Intent     string  // property_set / service_invoke / query_status
	Confidence float64

	// 实体槽位
	Slots map[string]any // device_name, property_id, service_id, value, action

	// 解析结果
	TargetDevice *model.Device
	ThingModel   *model.ThingModel
	VoiceModule  *model.ThingModelModule
	AllDevices   []*model.Device

	// 输出
	Result  *model.VoiceResult
	Aborted bool // 节点可设置此标志提前终止管道
}

// Pipeline 流程管道
type Pipeline struct {
	nodes []Node
}

// NewPipeline 创建管道
func NewPipeline(nodes ...Node) *Pipeline {
	return &Pipeline{nodes: nodes}
}

// Run 按顺序执行所有节点
func (p *Pipeline) Run(ctx context.Context, pc *PipelineContext) *model.VoiceResult {
	for _, node := range p.nodes {
		if pc.Aborted {
			break
		}
		if err := node.Process(ctx, pc); err != nil {
			return &model.VoiceResult{Success: false, Message: err.Error()}
		}
	}
	if pc.Result != nil {
		return pc.Result
	}
	return &model.VoiceResult{Success: false, Message: "管道未产生结果"}
}
