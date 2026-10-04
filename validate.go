// Package formvalidate 是设置面板校验：必填、范围、跨字段规则与错误排序。
// 缺陷：必填没做、范围端点用开区间、跨字段规则没做、结果按规则表顺序给、解析失败静默跳过、
// 每条规则扫全部字段。
package formvalidate

import (
	"errors"
	"strconv"
	"strings"
)

// ErrBadSchema 表单定义不合法。
var ErrBadSchema = errors.New("表单定义不合法")

// Field 是一个设置项。
type Field struct {
	Name     string
	Required bool
}

// Rule 是一条校验规则。Kind 为 range 时比较 Min 与 Max，为 le 或 ge 时与 Other 字段比较。
type Rule struct {
	ID       string
	Field    string
	Kind     string
	Priority int
	Min, Max float64
	Other    string
}

// Error 是一条校验错误。
type Error struct {
	Field string
	Rule  string
}

// Validator 是设置面板校验器。
type Validator struct {
	fields  []Field
	rules   []Rule
	values  map[string]string
	scanned int
}

// New 建校验器。
func New(fields []Field, rules []Rule) (*Validator, error) {
	known := map[string]bool{}
	for _, field := range fields {
		if field.Name == "" || known[field.Name] {
			return nil, ErrBadSchema
		}
		known[field.Name] = true
	}
	ids := map[string]bool{}
	for _, rule := range rules {
		if rule.ID == "" || ids[rule.ID] || !known[rule.Field] {
			return nil, ErrBadSchema
		}
		ids[rule.ID] = true
	}
	return &Validator{fields: fields, rules: rules, values: map[string]string{}}, nil
}

// Set 写入一个设置项的值。
func (v *Validator) Set(name, value string) bool {
	for _, field := range v.fields {
		if field.Name == name {
			v.values[name] = value
			return true
		}
	}
	return false
}

// Validate 返回当前的校验错误，按字段顺序、规则优先级、规则 ID 升序排列。
func (v *Validator) Validate() []Error {
	out := []Error{}
	for _, rule := range v.rules {
		for _, field := range v.fields {
			v.scanned++
			if field.Name != rule.Field {
				continue
			}
			value := strings.TrimSpace(v.values[field.Name])
			if rule.Kind != "range" {
				continue
			}
			number, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			if number <= rule.Min || number >= rule.Max {
				out = append(out, Error{Field: field.Name, Rule: rule.ID})
			}
		}
	}
	return out
}

// Scanned 是累计扫描的规则与字段配对数（规模观测）。
func (v *Validator) Scanned() int { return v.scanned }
