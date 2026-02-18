package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/mqtt"
)

// mockNode 用于测试 Pipeline 引擎
type mockNode struct {
	name    string
	fn      func(ctx context.Context, pc *mqtt.PipelineContext) error
}

func (n *mockNode) Name() string { return n.name }
func (n *mockNode) Process(ctx context.Context, pc *mqtt.PipelineContext) error {
	return n.fn(ctx, pc)
}

// ============ Pipeline 引擎测试 ============

func TestPipeline_Run_Success(t *testing.T) {
	p := mqtt.NewPipeline(
		&mockNode{"step1", func(_ context.Context, pc *mqtt.PipelineContext) error {
			pc.Intent = "test"
			return nil
		}},
		&mockNode{"step2", func(_ context.Context, pc *mqtt.PipelineContext) error {
			pc.Result = &model.VoiceResult{Success: true, Message: "ok"}
			return nil
		}},
	)

	result := p.Run(context.Background(), &mqtt.PipelineContext{RawText: "hello"})
	if !result.Success || result.Message != "ok" {
		t.Fatalf("expected success, got: %+v", result)
	}
}

func TestPipeline_Run_Abort(t *testing.T) {
	called := false
	p := mqtt.NewPipeline(
		&mockNode{"abort", func(_ context.Context, pc *mqtt.PipelineContext) error {
			pc.Aborted = true
			pc.Result = &model.VoiceResult{Success: false, Message: "aborted"}
			return nil
		}},
		&mockNode{"skip", func(_ context.Context, pc *mqtt.PipelineContext) error {
			called = true
			return nil
		}},
	)

	result := p.Run(context.Background(), &mqtt.PipelineContext{})
	if result.Success || result.Message != "aborted" {
		t.Fatalf("expected abort result, got: %+v", result)
	}
	if called {
		t.Fatal("second node should not be called after abort")
	}
}

func TestPipeline_Run_Error(t *testing.T) {
	p := mqtt.NewPipeline(
		&mockNode{"err", func(_ context.Context, pc *mqtt.PipelineContext) error {
			return errors.New("node error")
		}},
	)

	result := p.Run(context.Background(), &mqtt.PipelineContext{})
	if result.Success {
		t.Fatal("expected failure on error")
	}
	if result.Message != "node error" {
		t.Fatalf("expected error message, got: %s", result.Message)
	}
}

func TestPipeline_Run_NoResult(t *testing.T) {
	p := mqtt.NewPipeline(
		&mockNode{"noop", func(_ context.Context, pc *mqtt.PipelineContext) error {
			return nil
		}},
	)

	result := p.Run(context.Background(), &mqtt.PipelineContext{})
	if result.Success {
		t.Fatal("expected failure when no result produced")
	}
}

// ============ PreprocessNode 测试 ============

func TestPreprocessNode(t *testing.T) {
	node := &mqtt.PreprocessNode{}
	pc := &mqtt.PipelineContext{RawText: "  打开 客厅灯  "}

	if err := node.Process(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	if pc.RawText != "打开 客厅灯" {
		t.Fatalf("expected trimmed text, got: %q", pc.RawText)
	}
	if len(pc.Tokens) != 2 || pc.Tokens[0] != "打开" || pc.Tokens[1] != "客厅灯" {
		t.Fatalf("unexpected tokens: %v", pc.Tokens)
	}
	if pc.Slots == nil {
		t.Fatal("slots should be initialized")
	}
}

// ============ IntentClassifyNode 测试 ============

func TestIntentClassify_PropertySet(t *testing.T) {
	node := &mqtt.IntentClassifyNode{}
	tests := []struct {
		text   string
		intent string
	}{
		{"打开客厅灯", "property_set"},
		{"关闭空调", "property_set"},
		{"亮度调到80", "property_set"},
		{"温度调高一点", "property_set"},
		{"调低音量", "property_set"},
	}

	for _, tt := range tests {
		pc := &mqtt.PipelineContext{RawText: tt.text, Slots: make(map[string]any)}
		if err := node.Process(context.Background(), pc); err != nil {
			t.Fatalf("text=%q: %v", tt.text, err)
		}
		if pc.Intent != tt.intent {
			t.Errorf("text=%q: expected intent %q, got %q", tt.text, tt.intent, pc.Intent)
		}
	}
}

func TestIntentClassify_ServiceInvoke(t *testing.T) {
	node := &mqtt.IntentClassifyNode{}
	pc := &mqtt.PipelineContext{RawText: "重启设备", Slots: make(map[string]any)}

	if err := node.Process(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	if pc.Intent != "service_invoke" {
		t.Fatalf("expected service_invoke, got: %s", pc.Intent)
	}
}

func TestIntentClassify_QueryStatus(t *testing.T) {
	node := &mqtt.IntentClassifyNode{}
	pc := &mqtt.PipelineContext{RawText: "当前温度是多少", Slots: make(map[string]any)}

	if err := node.Process(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	if pc.Intent != "query_status" {
		t.Fatalf("expected query_status, got: %s", pc.Intent)
	}
}

func TestIntentClassify_Unknown(t *testing.T) {
	node := &mqtt.IntentClassifyNode{}
	pc := &mqtt.PipelineContext{RawText: "你好世界", Slots: make(map[string]any)}

	if err := node.Process(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	if !pc.Aborted {
		t.Fatal("expected abort on unknown intent")
	}
	if pc.Result == nil || pc.Result.Success {
		t.Fatal("expected failure result")
	}
}

func TestIntentClassify_CustomPatterns(t *testing.T) {
	node := &mqtt.IntentClassifyNode{
		Patterns: map[string][]mqtt.PatternRule{
			"custom_intent": {
				{Keywords: []string{"自定义"}, Score: 10},
			},
		},
	}
	pc := &mqtt.PipelineContext{RawText: "自定义指令", Slots: make(map[string]any)}

	if err := node.Process(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	if pc.Intent != "custom_intent" {
		t.Fatalf("expected custom_intent, got: %s", pc.Intent)
	}
}

// ============ SlotValidateNode 测试 ============

func TestSlotValidate_PropertySet_OK(t *testing.T) {
	node := &mqtt.SlotValidateNode{}
	pc := &mqtt.PipelineContext{
		Intent:       "property_set",
		TargetDevice: &model.Device{ID: "d1"},
		Slots:        map[string]any{"property_id": "brightness", "value": 80},
	}

	if err := node.Process(context.Background(), pc); err != nil {
		t.Fatal(err)
	}
	if pc.Aborted {
		t.Fatal("should not abort with valid slots")
	}
}

func TestSlotValidate_PropertySet_MissingValue(t *testing.T) {
	node := &mqtt.SlotValidateNode{}
	pc := &mqtt.PipelineContext{
		Intent:       "property_set",
		TargetDevice: &model.Device{ID: "d1"},
		Slots:        map[string]any{"property_id": "brightness"},
	}

	node.Process(context.Background(), pc)
	if !pc.Aborted {
		t.Fatal("should abort when value is missing")
	}
}

func TestSlotValidate_NoDevice(t *testing.T) {
	node := &mqtt.SlotValidateNode{}
	pc := &mqtt.PipelineContext{
		Intent: "property_set",
		Slots:  map[string]any{"property_id": "x", "value": 1},
	}

	node.Process(context.Background(), pc)
	if !pc.Aborted {
		t.Fatal("should abort when no target device")
	}
}

func TestSlotValidate_ServiceInvoke_MissingServiceID(t *testing.T) {
	node := &mqtt.SlotValidateNode{}
	pc := &mqtt.PipelineContext{
		Intent:       "service_invoke",
		TargetDevice: &model.Device{ID: "d1"},
		Slots:        map[string]any{},
	}

	node.Process(context.Background(), pc)
	if !pc.Aborted {
		t.Fatal("should abort when service_id is missing")
	}
}

func TestSlotValidate_QueryStatus_MissingPropertyID(t *testing.T) {
	node := &mqtt.SlotValidateNode{}
	pc := &mqtt.PipelineContext{
		Intent:       "query_status",
		TargetDevice: &model.Device{ID: "d1"},
		Slots:        map[string]any{},
	}

	node.Process(context.Background(), pc)
	if !pc.Aborted {
		t.Fatal("should abort when property_id is missing")
	}
}
