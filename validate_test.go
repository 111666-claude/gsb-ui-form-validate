package formvalidate

import (
	"strconv"
	"testing"
)

func TestNewRejectsDuplicateFieldName(t *testing.T) {
	fields := []Field{{Name: "a"}, {Name: "a"}}
	if _, err := New(fields, nil); err != ErrBadSchema {
		t.Fatal("字段名重复应该报 ErrBadSchema")
	}
}

func TestNewRejectsUnknownRuleField(t *testing.T) {
	fields := []Field{{Name: "a"}}
	rules := []Rule{{ID: "r1", Field: "b", Kind: "range", Max: 10}}
	if _, err := New(fields, rules); err != ErrBadSchema {
		t.Fatal("规则指向不存在的字段应该报 ErrBadSchema")
	}
}

func TestSetRejectsUnknownField(t *testing.T) {
	v, err := New([]Field{{Name: "a"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.Set("b", "1") {
		t.Fatal("写入不存在的字段应该返回 false")
	}
}

func TestRangeInsideIsClean(t *testing.T) {
	fields := []Field{{Name: "a"}}
	rules := []Rule{{ID: "r1", Field: "a", Kind: "range", Min: 0, Max: 100}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("a", "50")
	if got := v.Validate(); len(got) != 0 {
		t.Fatalf("区间内的值不该报错，得到 %v", got)
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	v, err := New([]Field{{Name: "a"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}

func TestRequiredEmptyReportsAndSkipsRules(t *testing.T) {
	fields := []Field{{Name: "a", Required: true}}
	rules := []Rule{{ID: "r1", Field: "a", Kind: "range", Min: 0, Max: 10}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("a", "   ")
	got := v.Validate()
	if len(got) != 1 || got[0] != (Error{Field: "a", Rule: "required"}) {
		t.Fatalf("必填空白应该只报一条 required，得到 %v", got)
	}
}

func TestRequiredFilledStillRunsRules(t *testing.T) {
	fields := []Field{{Name: "a", Required: true}}
	rules := []Rule{{ID: "r1", Field: "a", Kind: "range", Min: 0, Max: 10}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("a", "99")
	got := v.Validate()
	if len(got) != 1 || got[0].Rule != "r1" {
		t.Fatalf("必填已填应该继续判 range，得到 %v", got)
	}
}

func TestRangeEndpointsAreInclusive(t *testing.T) {
	fields := []Field{{Name: "a"}}
	rules := []Rule{{ID: "r1", Field: "a", Kind: "range", Min: 0, Max: 100}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"0", "100"} {
		v.Set("a", value)
		if got := v.Validate(); len(got) != 0 {
			t.Fatalf("端点 %s 应该通过，得到 %v", value, got)
		}
	}
	for _, value := range []string{"-1", "101"} {
		v.Set("a", value)
		if got := v.Validate(); len(got) != 1 {
			t.Fatalf("越界 %s 应该报一条，得到 %v", value, got)
		}
	}
}

func TestRangeUnparseableIsError(t *testing.T) {
	fields := []Field{{Name: "a"}}
	rules := []Rule{{ID: "r1", Field: "a", Kind: "range", Min: 0, Max: 10}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("a", "abc")
	if got := v.Validate(); len(got) != 1 || got[0].Rule != "r1" {
		t.Fatalf("解析失败应该报一条错误，得到 %v", got)
	}
}

func TestCrossFieldLeAndGe(t *testing.T) {
	fields := []Field{{Name: "fps"}, {Name: "hz"}}
	rules := []Rule{
		{ID: "le", Field: "fps", Kind: "le", Other: "hz"},
		{ID: "ge", Field: "hz", Kind: "ge", Other: "fps"},
	}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("fps", "144")
	v.Set("hz", "60")
	got := v.Validate()
	if len(got) != 2 || got[0].Rule != "le" || got[1].Rule != "ge" {
		t.Fatalf("fps>hz 应该 le 与 ge 都报错，得到 %v", got)
	}
	v.Set("fps", "60")
	if got := v.Validate(); len(got) != 0 {
		t.Fatalf("相等应该通过，得到 %v", got)
	}
}

func TestCrossFieldSkipsWhenOtherMissingOrBad(t *testing.T) {
	fields := []Field{{Name: "fps"}, {Name: "hz"}}
	rules := []Rule{{ID: "le", Field: "fps", Kind: "le", Other: "hz"}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("fps", "144")
	if got := v.Validate(); len(got) != 0 {
		t.Fatalf("Other 没设值应该跳过，得到 %v", got)
	}
	v.Set("hz", "abc")
	if got := v.Validate(); len(got) != 0 {
		t.Fatalf("Other 解析失败应该跳过，得到 %v", got)
	}
}

func TestErrorsOrderedByFieldPriorityAndID(t *testing.T) {
	fields := []Field{{Name: "a"}, {Name: "b"}}
	rules := []Rule{
		{ID: "z", Field: "b", Kind: "range", Min: 0, Max: 10, Priority: 2},
		{ID: "y", Field: "b", Kind: "range", Min: 0, Max: 10, Priority: 1},
		{ID: "b2", Field: "a", Kind: "range", Min: 0, Max: 10},
		{ID: "b1", Field: "a", Kind: "range", Min: 0, Max: 10},
	}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("a", "99")
	v.Set("b", "99")
	got := v.Validate()
	want := []Error{
		{Field: "a", Rule: "b1"},
		{Field: "a", Rule: "b2"},
		{Field: "b", Rule: "y"},
		{Field: "b", Rule: "z"},
	}
	if len(got) != len(want) {
		t.Fatalf("错误数不对，得到 %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 条应该是 %v，得到 %v", i, want[i], got[i])
		}
	}
}

func TestValidateIsRepeatableAndDoesNotMutate(t *testing.T) {
	fields := []Field{{Name: "a", Required: true}, {Name: "b"}}
	rules := []Rule{{ID: "r1", Field: "b", Kind: "range", Min: 0, Max: 10}}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("b", "99")
	first := v.Validate()
	second := v.Validate()
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("重复校验应该一致，得到 %v 与 %v", first, second)
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("重复校验第 %d 条不一致：%v 与 %v", i, first[i], second[i])
		}
	}
	v.Set("b", "5")
	if got := v.Validate(); len(got) != 1 || got[0].Rule != "required" {
		t.Fatalf("改值后应该只剩 required，得到 %v", got)
	}
}

func TestScannedBoundedByFieldsPlusRules(t *testing.T) {
	fields := []Field{}
	rules := []Rule{}
	for i := 0; i < 50; i++ {
		name := "f-" + strconv.Itoa(i)
		fields = append(fields, Field{Name: name})
		rules = append(rules, Rule{ID: "r-" + strconv.Itoa(i), Field: name, Kind: "range", Min: 0, Max: 100})
	}
	v, err := New(fields, rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range fields {
		v.Set(field.Name, "50")
	}
	v.Validate()
	perRun := v.Scanned()
	if perRun > len(fields)+len(rules) {
		t.Fatalf("单次扫描 %d 超过字段数加规则数 %d", perRun, len(fields)+len(rules))
	}
	v.Validate()
	if v.Scanned() != 2*perRun {
		t.Fatalf("Scanned 应该按次累计，得到 %d", v.Scanned())
	}
}
