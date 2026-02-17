package mqtt

import (
	"fmt"
	"math"

	"github.com/ldchengyi/linkflow/internal/model"
)

// ValidateResult 校验结果
type ValidateResult struct {
	Valid   bool              // 所有字段是否都合法
	Payload map[string]interface{} // 原始 payload（全量保留）
	Errors  map[string]string      // 不合法字段 → 原因
}

// ValidatePayload 根据物模型属性定义校验 payload
func ValidatePayload(payload map[string]interface{}, properties []model.Property) ValidateResult {
	result := ValidateResult{
		Valid:   true,
		Payload: payload,
		Errors:  make(map[string]string),
	}

	// 构建属性 ID → Property 的映射
	propMap := make(map[string]model.Property, len(properties))
	for _, p := range properties {
		propMap[p.ID] = p
	}

	for key, value := range payload {
		prop, exists := propMap[key]
		if !exists {
			result.Errors[key] = "未知属性"
			result.Valid = false
			continue
		}

		if err := validateValue(value, prop); err != "" {
			result.Errors[key] = err
			result.Valid = false
		}
	}

	return result
}

// validateValue 校验单个值是否符合属性定义
func validateValue(value interface{}, prop model.Property) string {
	switch prop.DataType {
	case "int":
		return validateInt(value, prop)
	case "float":
		return validateFloat(value, prop)
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Sprintf("类型错误: 期望 bool, 实际 %T", value)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Sprintf("类型错误: 期望 string, 实际 %T", value)
		}
	case "enum":
		return validateEnum(value, prop)
	default:
		return fmt.Sprintf("未知数据类型: %s", prop.DataType)
	}
	return ""
}

// validateInt 校验整数值
func validateInt(value interface{}, prop model.Property) string {
	// JSON 数字统一解析为 float64
	num, ok := value.(float64)
	if !ok {
		return fmt.Sprintf("类型错误: 期望 int, 实际 %T", value)
	}
	if num != math.Trunc(num) {
		return fmt.Sprintf("类型错误: 期望整数, 实际 %v", num)
	}
	return checkRange(num, prop)
}

// validateFloat 校验浮点值
func validateFloat(value interface{}, prop model.Property) string {
	num, ok := value.(float64)
	if !ok {
		return fmt.Sprintf("类型错误: 期望 float, 实际 %T", value)
	}
	return checkRange(num, prop)
}

// checkRange 检查数值范围
func checkRange(num float64, prop model.Property) string {
	if prop.Min != nil && num < *prop.Min {
		return fmt.Sprintf("低于最小值: %v < %v", num, *prop.Min)
	}
	if prop.Max != nil && num > *prop.Max {
		return fmt.Sprintf("超出最大值: %v > %v", num, *prop.Max)
	}
	return ""
}

// validateEnum 校验枚举值
func validateEnum(value interface{}, prop model.Property) string {
	num, ok := value.(float64)
	if !ok {
		return fmt.Sprintf("类型错误: 期望 enum(int), 实际 %T", value)
	}
	intVal := int(num)
	for _, ev := range prop.EnumValues {
		if ev.Value == intVal {
			return ""
		}
	}
	return fmt.Sprintf("枚举值不合法: %v", intVal)
}
