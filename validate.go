// Package formvalidate 是设置面板校验：必填、范围、跨字段规则与错误排序。
package formvalidate

import (
	"errors"
	"sort"
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
	index   map[string]int
	rules   map[string][]Rule
	values  map[string]string
	set     map[string]bool
	dirty   bool
	cached  []Error
	scanned int
}

// New 建校验器。
func New(fields []Field, rules []Rule) (*Validator, error) {
	index := map[string]int{}
	for _, field := range fields {
		if field.Name == "" {
			return nil, ErrBadSchema
		}
		if _, ok := index[field.Name]; ok {
			return nil, ErrBadSchema
		}
		index[field.Name] = len(index)
	}
	ids := map[string]bool{}
	byField := map[string][]Rule{}
	for _, rule := range rules {
		if rule.ID == "" || ids[rule.ID] {
			return nil, ErrBadSchema
		}
		if _, ok := index[rule.Field]; !ok {
			return nil, ErrBadSchema
		}
		switch rule.Kind {
		case "range":
		case "le", "ge":
			if _, ok := index[rule.Other]; !ok {
				return nil, ErrBadSchema
			}
		default:
			return nil, ErrBadSchema
		}
		ids[rule.ID] = true
		byField[rule.Field] = append(byField[rule.Field], rule)
	}
	for _, list := range byField {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Priority != list[j].Priority {
				return list[i].Priority < list[j].Priority
			}
			return list[i].ID < list[j].ID
		})
	}
	return &Validator{
		fields: fields,
		index:  index,
		rules:  byField,
		values: map[string]string{},
		set:    map[string]bool{},
		dirty:  true,
	}, nil
}

// Set 写入一个设置项的值。
func (v *Validator) Set(name, value string) bool {
	if _, ok := v.index[name]; !ok {
		return false
	}
	v.values[name] = value
	v.set[name] = true
	v.dirty = true
	return true
}

// Validate 返回当前的校验错误，按字段顺序、规则优先级、规则 ID 升序排列。
// 表单自上次校验后没有变化时直接返回缓存结果。
func (v *Validator) Validate() []Error {
	if !v.dirty {
		return append([]Error{}, v.cached...)
	}
	out := []Error{}
	for _, field := range v.fields {
		fieldRules := v.rules[field.Name]
		if !field.Required && len(fieldRules) == 0 {
			continue
		}
		v.scanned++
		value := strings.TrimSpace(v.values[field.Name])
		if field.Required && value == "" {
			out = append(out, Error{Field: field.Name, Rule: "required"})
			continue
		}
		for _, rule := range fieldRules {
			v.scanned++
			if v.violated(rule, value) {
				out = append(out, Error{Field: field.Name, Rule: rule.ID})
			}
		}
	}
	v.cached = append([]Error{}, out...)
	v.dirty = false
	return out
}

// violated 判断一条规则是否对当前值成立（成立即报错）。返回 false 也可能表示规则被跳过。
func (v *Validator) violated(rule Rule, value string) bool {
	number, err := strconv.ParseFloat(value, 64)
	switch rule.Kind {
	case "range":
		if err != nil {
			return true
		}
		return number < rule.Min || number > rule.Max
	case "le", "ge":
		if err != nil {
			return true
		}
		if !v.set[rule.Other] {
			return false
		}
		other, err := strconv.ParseFloat(strings.TrimSpace(v.values[rule.Other]), 64)
		if err != nil {
			return false
		}
		if rule.Kind == "le" {
			return number > other
		}
		return number < other
	}
	return false
}

// Scanned 是累计扫描的规则与字段配对数（规模观测）。
func (v *Validator) Scanned() int { return v.scanned }
