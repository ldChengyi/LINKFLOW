package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
)

// ============ 意图模式定义（可配置化） ============

// PatternRule 关键词模式规则
type PatternRule struct {
	Keywords []string
	Score    int
}

// 默认意图模式表
var defaultIntentPatterns = map[string][]PatternRule{
	"property_set": {
		{Keywords: []string{"打开", "开启", "开灯", "启动"}, Score: 10},
		{Keywords: []string{"关闭", "关掉", "关灯", "停止"}, Score: 10},
		{Keywords: []string{"调到", "设为", "设置到", "设置为"}, Score: 10},
		{Keywords: []string{"调高", "增大", "调亮", "升高", "加大", "高一点", "亮一点", "大一点"}, Score: 8},
		{Keywords: []string{"调低", "减小", "调暗", "降低", "低一点", "暗一点", "小一点"}, Score: 8},
	},
	"service_invoke": {
		{Keywords: []string{"执行", "调用", "运行"}, Score: 8},
		{Keywords: []string{"重启", "重置", "复位"}, Score: 6},
	},
	"query_status": {
		{Keywords: []string{"多少", "什么", "查询", "当前", "状态", "是多少"}, Score: 10},
	},
}

var numRegex = regexp.MustCompile(`(\d+\.?\d*)`)

// ============ Node 1: 预处理 ============

type PreprocessNode struct{}

func (n *PreprocessNode) Name() string { return "preprocess" }

func (n *PreprocessNode) Process(_ context.Context, pc *PipelineContext) error {
	pc.RawText = strings.TrimSpace(pc.RawText)
	pc.Slots = make(map[string]any)
	// 简单分词：按空格 + 保留原文用于子串匹配
	pc.Tokens = strings.Fields(pc.RawText)
	return nil
}

// ============ Node 2: 加载设备列表 ============

type DeviceLoadNode struct {
	broker *Broker
}

func (n *DeviceLoadNode) Name() string { return "device_load" }

func (n *DeviceLoadNode) Process(ctx context.Context, pc *PipelineContext) error {
	devices, err := n.broker.deviceRepo.ListByUserID(ctx, pc.UserID)
	if err != nil {
		return fmt.Errorf("获取设备列表失败")
	}
	pc.AllDevices = devices
	return nil
}

// ============ Node 3: 意图分类（模式打分） ============

type IntentClassifyNode struct {
	Patterns map[string][]PatternRule
}

func (n *IntentClassifyNode) Name() string { return "intent_classify" }

func (n *IntentClassifyNode) Process(_ context.Context, pc *PipelineContext) error {
	patterns := n.Patterns
	if patterns == nil {
		patterns = defaultIntentPatterns
	}

	var bestIntent string
	var bestScore int

	for intent, rules := range patterns {
		score := 0
		for _, rule := range rules {
			for _, kw := range rule.Keywords {
				if strings.Contains(pc.RawText, kw) {
					score += rule.Score
					break // 同一规则组只计一次
				}
			}
		}
		if score > bestScore {
			bestScore = score
			bestIntent = intent
		}
	}

	if bestIntent == "" {
		pc.Result = &model.VoiceResult{Success: false, Message: "无法识别指令意图: " + pc.RawText}
		pc.Aborted = true
		return nil
	}

	pc.Intent = bestIntent
	pc.Confidence = float64(bestScore)
	return nil
}

// ============ Node 4: 实体提取（设备 + 属性/服务 + 值） ============

type EntityExtractNode struct {
	broker *Broker
}

func (n *EntityExtractNode) Name() string { return "entity_extract" }

func (n *EntityExtractNode) Process(ctx context.Context, pc *PipelineContext) error {
	text := pc.RawText

	// 1. 最长匹配设备名
	var bestDevice *model.Device
	var bestLen int
	for _, d := range pc.AllDevices {
		if strings.Contains(text, d.Name) && len(d.Name) > bestLen {
			bestDevice = d
			bestLen = len(d.Name)
		}
	}

	// 2. 如果匹配到设备，在该设备上提取属性/服务
	if bestDevice != nil {
		if n.extractFromDevice(ctx, pc, bestDevice) {
			return nil
		}
	}

	// 3. 反向查找：通过属性名/服务名匹配设备
	n.extractByReverseLookup(ctx, pc)
	return nil
}

// extractFromDevice 在指定设备上提取属性/服务槽位
func (n *EntityExtractNode) extractFromDevice(ctx context.Context, pc *PipelineContext, device *model.Device) bool {
	if device.ModelID == nil {
		return false
	}
	tm, err := n.broker.thingModelRepo.GetByID(ctx, *device.ModelID)
	if err != nil {
		return false
	}
	vm := findVoiceModule(tm)
	if vm == nil {
		return false
	}

	pc.TargetDevice = device
	pc.ThingModel = tm
	pc.VoiceModule = vm

	text := pc.RawText

	// 属性名精确匹配 + 值提取
	if pc.Intent == "property_set" {
		if n.matchProperty(pc, tm, vm, text) {
			return true
		}
		// Bool 模糊匹配：仅一个 exposed bool 属性时生效
		if n.matchSingleBool(pc, tm, vm, text) {
			return true
		}
	}

	// 查询意图：匹配属性名
	if pc.Intent == "query_status" {
		for _, prop := range tm.Properties {
			if !slotExposed(prop.ID, vm.Config.ExposedProperties) {
				continue
			}
			if strings.Contains(text, prop.Name) || strings.Contains(text, prop.ID) {
				pc.Slots["property_id"] = prop.ID
				return true
			}
		}
	}

	// 服务匹配
	if pc.Intent == "service_invoke" {
		for _, svc := range tm.Services {
			if !slotExposed(svc.ID, vm.Config.ExposedServices) {
				continue
			}
			if strings.Contains(text, svc.Name) || strings.Contains(text, svc.ID) {
				pc.Slots["service_id"] = svc.ID
				return true
			}
		}
	}

	return false
}

// matchProperty 精确匹配属性名 + 提取值
func (n *EntityExtractNode) matchProperty(pc *PipelineContext, tm *model.ThingModel, vm *model.ThingModelModule, text string) bool {
	for _, prop := range tm.Properties {
		if prop.AccessMode != "rw" || !slotExposed(prop.ID, vm.Config.ExposedProperties) {
			continue
		}
		if !strings.Contains(text, prop.Name) && !strings.Contains(text, prop.ID) {
			continue
		}
		value := parseSlotValue(text, prop)
		if value != nil {
			pc.Slots["property_id"] = prop.ID
			pc.Slots["value"] = value
			return true
		}
	}
	return false
}

// matchSingleBool 仅一个 exposed bool 属性时的模糊匹配
func (n *EntityExtractNode) matchSingleBool(pc *PipelineContext, tm *model.ThingModel, vm *model.ThingModelModule, text string) bool {
	var bools []model.Property
	for _, prop := range tm.Properties {
		if prop.AccessMode == "rw" && prop.DataType == "bool" && slotExposed(prop.ID, vm.Config.ExposedProperties) {
			bools = append(bools, prop)
		}
	}
	if len(bools) != 1 {
		return false
	}
	if matchOn(text) {
		pc.Slots["property_id"] = bools[0].ID
		pc.Slots["value"] = true
		return true
	}
	if matchOff(text) {
		pc.Slots["property_id"] = bools[0].ID
		pc.Slots["value"] = false
		return true
	}
	return false
}

// extractByReverseLookup 通过属性名/服务名反向查找设备
func (n *EntityExtractNode) extractByReverseLookup(ctx context.Context, pc *PipelineContext) {
	text := pc.RawText

	type candidate struct {
		device *model.Device
		tm     *model.ThingModel
		vm     *model.ThingModelModule
		propID string
		value  any
	}
	var candidates []candidate

	for _, d := range pc.AllDevices {
		if d.ModelID == nil {
			continue
		}
		tm, err := n.broker.thingModelRepo.GetByID(ctx, *d.ModelID)
		if err != nil {
			continue
		}
		vm := findVoiceModule(tm)
		if vm == nil {
			continue
		}

		if pc.Intent == "property_set" || pc.Intent == "query_status" {
			for _, prop := range tm.Properties {
				if !slotExposed(prop.ID, vm.Config.ExposedProperties) {
					continue
				}
				if !strings.Contains(text, prop.Name) && !strings.Contains(text, prop.ID) {
					continue
				}
				if pc.Intent == "query_status" {
					candidates = append(candidates, candidate{d, tm, vm, prop.ID, nil})
					continue
				}
				if prop.AccessMode != "rw" {
					continue
				}
				if v := parseSlotValue(text, prop); v != nil {
					candidates = append(candidates, candidate{d, tm, vm, prop.ID, v})
				}
			}
		}

		if pc.Intent == "service_invoke" {
			for _, svc := range tm.Services {
				if !slotExposed(svc.ID, vm.Config.ExposedServices) {
					continue
				}
				if strings.Contains(text, svc.Name) || strings.Contains(text, svc.ID) {
					pc.TargetDevice = d
					pc.ThingModel = tm
					pc.VoiceModule = vm
					pc.Slots["service_id"] = svc.ID
					return
				}
			}
		}
	}

	// 唯一匹配才采纳，多个有歧义则放弃
	if len(candidates) == 1 {
		c := candidates[0]
		pc.TargetDevice = c.device
		pc.ThingModel = c.tm
		pc.VoiceModule = c.vm
		pc.Slots["property_id"] = c.propID
		if c.value != nil {
			pc.Slots["value"] = c.value
		}
	}
}

// ============ Node 5: 槽位校验 ============

type SlotValidateNode struct{}

func (n *SlotValidateNode) Name() string { return "slot_validate" }

func (n *SlotValidateNode) Process(_ context.Context, pc *PipelineContext) error {
	if pc.TargetDevice == nil {
		pc.Result = &model.VoiceResult{Success: false, Message: "未匹配到目标设备"}
		pc.Aborted = true
		return nil
	}

	switch pc.Intent {
	case "property_set":
		if pc.Slots["property_id"] == nil || pc.Slots["value"] == nil {
			pc.Result = &model.VoiceResult{Success: false, Message: "无法识别要设置的属性或值"}
			pc.Aborted = true
		}
	case "service_invoke":
		if pc.Slots["service_id"] == nil {
			pc.Result = &model.VoiceResult{Success: false, Message: "无法识别要调用的服务"}
			pc.Aborted = true
		}
	case "query_status":
		if pc.Slots["property_id"] == nil {
			pc.Result = &model.VoiceResult{Success: false, Message: "无法识别要查询的属性"}
			pc.Aborted = true
		}
	}
	return nil
}

// ============ Node 6: 执行动作 ============

type ExecuteNode struct {
	broker *Broker
}

func (n *ExecuteNode) Name() string { return "execute" }

func (n *ExecuteNode) Process(ctx context.Context, pc *PipelineContext) error {
	switch pc.Intent {
	case "property_set":
		pc.Result = n.execPropertySet(ctx, pc)
	case "service_invoke":
		pc.Result = n.execServiceInvoke(pc)
	case "query_status":
		pc.Result = n.execQueryStatus(ctx, pc)
	}
	return nil
}

func (n *ExecuteNode) execPropertySet(ctx context.Context, pc *PipelineContext) *model.VoiceResult {
	deviceID := pc.TargetDevice.ID
	propID := pc.Slots["property_id"].(string)
	value := pc.Slots["value"]

	// 处理相对值（调高/调低）
	if actionMap, ok := value.(map[string]string); ok {
		resolved, err := resolveRelative(ctx, n.broker, deviceID, propID, pc.ThingModel, actionMap["_action"])
		if err != nil {
			return &model.VoiceResult{Success: false, Message: err.Error()}
		}
		value = resolved
	}

	topic := fmt.Sprintf("devices/%s/telemetry/down", deviceID)
	payload, _ := json.Marshal(map[string]any{propID: value})

	if err := n.broker.Publish(topic, payload, false, 1); err != nil {
		return &model.VoiceResult{Success: false, Message: "指令下发失败"}
	}

	// 同步写入 device_data
	if pc.UserID != "" {
		merged := map[string]any{propID: value}
		if existing, err := n.broker.deviceDataRepo.GetLatestData(ctx, deviceID); err == nil && existing != nil {
			for k, v := range existing.Payload {
				if _, set := merged[k]; !set {
					merged[k] = v
				}
			}
		}
		storeTopic := fmt.Sprintf("devices/%s/telemetry/up", deviceID)
		_ = n.broker.deviceDataRepo.InsertTelemetry(ctx, deviceID, pc.UserID, storeTopic, merged, 1, true, nil)
	}

	action := fmt.Sprintf("设置 %s.%s = %v", pc.TargetDevice.Name, propID, value)
	logger.Log.Infof("Voice executed: %s", action)
	return &model.VoiceResult{Success: true, Message: "指令已执行", Action: action}
}

func (n *ExecuteNode) execServiceInvoke(pc *PipelineContext) *model.VoiceResult {
	deviceID := pc.TargetDevice.ID
	svcID := pc.Slots["service_id"].(string)

	topic := fmt.Sprintf("devices/%s/service/invoke", deviceID)
	payload, _ := json.Marshal(map[string]any{
		"id":      fmt.Sprintf("voice_%d", time.Now().UnixMilli()),
		"service": svcID,
		"params":  map[string]any{},
	})

	if err := n.broker.Publish(topic, payload, false, 1); err != nil {
		return &model.VoiceResult{Success: false, Message: "服务调用失败"}
	}

	action := fmt.Sprintf("调用 %s.%s", pc.TargetDevice.Name, svcID)
	logger.Log.Infof("Voice executed: %s", action)
	return &model.VoiceResult{Success: true, Message: "服务已调用", Action: action}
}

func (n *ExecuteNode) execQueryStatus(ctx context.Context, pc *PipelineContext) *model.VoiceResult {
	deviceID := pc.TargetDevice.ID
	propID := pc.Slots["property_id"].(string)

	data, err := n.broker.deviceDataRepo.GetLatestData(ctx, deviceID)
	if err != nil || data == nil {
		return &model.VoiceResult{Success: false, Message: "无法获取设备数据"}
	}

	val, ok := data.Payload[propID]
	if !ok {
		return &model.VoiceResult{Success: false, Message: fmt.Sprintf("属性 %s 无数据", propID)}
	}

	action := fmt.Sprintf("查询 %s.%s = %v", pc.TargetDevice.Name, propID, val)
	return &model.VoiceResult{Success: true, Message: fmt.Sprintf("%s 当前值: %v", propID, val), Action: action}
}

// ============ 工具函数 ============

var (
	onKeywords       = []string{"打开", "开启", "开灯", "启动", "开"}
	offKeywords      = []string{"关闭", "关掉", "关灯", "停止", "关"}
	increaseKeywords = []string{"调高", "增大", "调亮", "升高", "加大", "高一点", "亮一点", "大一点"}
	decreaseKeywords = []string{"调低", "减小", "调暗", "降低", "低一点", "暗一点", "小一点"}
)

func matchOn(text string) bool {
	for _, w := range onKeywords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func matchOff(text string) bool {
	for _, w := range offKeywords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

// parseSlotValue 根据属性类型从文本提取值
func parseSlotValue(text string, prop model.Property) any {
	switch prop.DataType {
	case "bool":
		if matchOn(text) {
			return true
		}
		if matchOff(text) {
			return false
		}
	case "int":
		if v, ok := extractNum(text); ok {
			iv := int(v)
			if prop.Min != nil && float64(iv) < *prop.Min {
				iv = int(*prop.Min)
			}
			if prop.Max != nil && float64(iv) > *prop.Max {
				iv = int(*prop.Max)
			}
			return iv
		}
		if matchInc(text) {
			return map[string]string{"_action": "increase"}
		}
		if matchDec(text) {
			return map[string]string{"_action": "decrease"}
		}
	case "float":
		if v, ok := extractNum(text); ok {
			if prop.Min != nil && v < *prop.Min {
				v = *prop.Min
			}
			if prop.Max != nil && v > *prop.Max {
				v = *prop.Max
			}
			return v
		}
		if matchInc(text) {
			return map[string]string{"_action": "increase"}
		}
		if matchDec(text) {
			return map[string]string{"_action": "decrease"}
		}
	case "enum":
		for _, ev := range prop.EnumValues {
			if strings.Contains(text, ev.Label) {
				return ev.Value
			}
		}
	}
	return nil
}

func matchInc(text string) bool {
	for _, w := range increaseKeywords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func matchDec(text string) bool {
	for _, w := range decreaseKeywords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func extractNum(text string) (float64, bool) {
	m := numRegex.FindStringSubmatch(text)
	if len(m) < 2 {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	return v, err == nil
}

func findVoiceModule(tm *model.ThingModel) *model.ThingModelModule {
	for i, m := range tm.Modules {
		if m.ID == "voice" {
			return &tm.Modules[i]
		}
	}
	return nil
}

func slotExposed(id string, exposed []string) bool {
	for _, e := range exposed {
		if e == id {
			return true
		}
	}
	return false
}

// resolveRelative 解析相对值（调高/调低）
func resolveRelative(ctx context.Context, broker *Broker, deviceID, propID string, tm *model.ThingModel, action string) (any, error) {
	data, err := broker.deviceDataRepo.GetLatestData(ctx, deviceID)
	if err != nil || data == nil {
		return nil, fmt.Errorf("无法获取设备当前数据")
	}
	currentVal, ok := data.Payload[propID]
	if !ok {
		return nil, fmt.Errorf("属性 %s 无当前值", propID)
	}

	var prop *model.Property
	for i, p := range tm.Properties {
		if p.ID == propID {
			prop = &tm.Properties[i]
			break
		}
	}
	if prop == nil {
		return nil, fmt.Errorf("属性 %s 未定义", propID)
	}

	step := 10.0
	if prop.Step != nil {
		step = *prop.Step
	}

	current, _ := toFloat(currentVal)
	newVal := current + step
	if action == "decrease" {
		newVal = current - step
	}
	if prop.Min != nil && newVal < *prop.Min {
		newVal = *prop.Min
	}
	if prop.Max != nil && newVal > *prop.Max {
		newVal = *prop.Max
	}

	if prop.DataType == "int" {
		return int(newVal), nil
	}
	return newVal, nil
}

func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}