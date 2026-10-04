package formvalidate

import "testing"

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
