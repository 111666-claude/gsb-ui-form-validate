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

// parsedValue 是某个字段值的解析缓存，gen 与 Validator.gen 一致时有效。
type parsedValue struct {
	number float64
	ok     bool
	gen    uint64
}

// indexedRule 是把 Other 解析成字段下标后的规则。
type indexedRule struct {
	Rule
	other int
}

// Validator 是设置面板校验器。
type Validator struct {
	fields  []Field
	index   map[string]int
	byField [][]indexedRule
	active  []int
	values  []string
	set     []bool
	scanned int
	parsed  []parsedValue
	gen     uint64
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
		switch rule.Kind {
		case "range":
		case "le", "ge":
			if !known[rule.Other] {
				return nil, ErrBadSchema
			}
		default:
			return nil, ErrBadSchema
		}
		ids[rule.ID] = true
	}
	v := &Validator{
		fields:  fields,
		index:   make(map[string]int, len(fields)),
		byField: make([][]indexedRule, len(fields)),
		values:  make([]string, len(fields)),
		set:     make([]bool, len(fields)),
		parsed:  make([]parsedValue, len(fields)),
	}
	for i, field := range fields {
		v.index[field.Name] = i
	}
	for _, rule := range rules {
		i := v.index[rule.Field]
		bound := indexedRule{Rule: rule, other: -1}
		if rule.Kind == "le" || rule.Kind == "ge" {
			bound.other = v.index[rule.Other]
		}
		v.byField[i] = append(v.byField[i], bound)
	}
	for i := range v.byField {
		rulesForField := v.byField[i]
		sort.SliceStable(rulesForField, func(a, b int) bool {
			if rulesForField[a].Priority != rulesForField[b].Priority {
				return rulesForField[a].Priority < rulesForField[b].Priority
			}
			return rulesForField[a].ID < rulesForField[b].ID
		})
		if len(rulesForField) > 0 || fields[i].Required {
			v.active = append(v.active, i)
		}
	}
	return v, nil
}

// Set 写入一个设置项的值。
func (v *Validator) Set(name, value string) bool {
	i, ok := v.index[name]
	if !ok {
		return false
	}
	v.values[i] = value
	v.set[i] = true
	return true
}

// number 解析字段值，同一次 Validate 内只解析一次。
func (v *Validator) number(i int) (float64, bool) {
	entry := &v.parsed[i]
	if entry.gen != v.gen {
		raw := strings.TrimSpace(v.values[i])
		number, err := strconv.ParseFloat(raw, 64)
		entry.number = number
		entry.ok = err == nil
		entry.gen = v.gen
	}
	return entry.number, entry.ok
}

// Validate 返回当前的校验错误，按字段顺序、规则优先级、规则 ID 升序排列。
func (v *Validator) Validate() []Error {
	v.gen++
	out := []Error{}
	for _, i := range v.active {
		field := v.fields[i]
		trimmed := strings.TrimSpace(v.values[i])
		if field.Required && trimmed == "" {
			out = append(out, Error{Field: field.Name, Rule: "required"})
			continue
		}
		for _, rule := range v.byField[i] {
			v.scanned++
			switch rule.Kind {
			case "range":
				number, ok := v.number(i)
				if !ok || number < rule.Min || number > rule.Max {
					out = append(out, Error{Field: field.Name, Rule: rule.ID})
				}
			case "le", "ge":
				if !v.set[rule.other] {
					continue
				}
				other, err := strconv.ParseFloat(strings.TrimSpace(v.values[rule.other]), 64)
				if err != nil {
					continue
				}
				number, ok := v.number(i)
				if !ok {
					out = append(out, Error{Field: field.Name, Rule: rule.ID})
					continue
				}
				if rule.Kind == "le" && number > other {
					out = append(out, Error{Field: field.Name, Rule: rule.ID})
				}
				if rule.Kind == "ge" && number < other {
					out = append(out, Error{Field: field.Name, Rule: rule.ID})
				}
			}
		}
	}
	return out
}

// Scanned 是累计扫描的规则与字段配对数（规模观测）。
func (v *Validator) Scanned() int { return v.scanned }
