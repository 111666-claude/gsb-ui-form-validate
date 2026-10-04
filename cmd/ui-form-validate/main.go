// Command ui-form-validate 跑设置面板校验样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/formvalidate"
)

func fieldNames(errors []formvalidate.Error) string {
	names := []string{}
	for _, item := range errors {
		names = append(names, item.Field)
	}
	return strings.Join(names, ",")
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ui-form-validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "required", "required / bound / cross / order / parse / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample == "required" {
		v, err := formvalidate.New([]formvalidate.Field{{Name: "volume", Required: true}}, nil)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "errors=%d\n", len(v.Validate()))
		return 0
	}
	if *sample == "bound" {
		fields := []formvalidate.Field{{Name: "fps"}}
		rules := []formvalidate.Rule{{ID: "r1", Field: "fps", Kind: "range", Min: 0, Max: 100}}
		v, err := formvalidate.New(fields, rules)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		v.Set("fps", "100")
		fmt.Fprintf(stdout, "errors=%d\n", len(v.Validate()))
		return 0
	}
	if *sample == "cross" {
		fields := []formvalidate.Field{{Name: "fps"}, {Name: "hz"}}
		rules := []formvalidate.Rule{{ID: "c1", Field: "fps", Kind: "le", Other: "hz"}}
		v, err := formvalidate.New(fields, rules)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		v.Set("fps", "144")
		v.Set("hz", "60")
		fmt.Fprintf(stdout, "errors=%d\n", len(v.Validate()))
		return 0
	}
	if *sample == "order" {
		fields := []formvalidate.Field{{Name: "a"}, {Name: "b"}}
		rules := []formvalidate.Rule{
			{ID: "r-b", Field: "b", Kind: "range", Min: 0, Max: 10},
			{ID: "r-a", Field: "a", Kind: "range", Min: 0, Max: 10},
		}
		v, err := formvalidate.New(fields, rules)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		v.Set("a", "99")
		v.Set("b", "99")
		fmt.Fprintf(stdout, "order=%s\n", fieldNames(v.Validate()))
		return 0
	}
	if *sample == "parse" {
		fields := []formvalidate.Field{{Name: "a"}}
		rules := []formvalidate.Rule{{ID: "r1", Field: "a", Kind: "range", Min: 0, Max: 10}}
		v, err := formvalidate.New(fields, rules)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		v.Set("a", "abc")
		fmt.Fprintf(stdout, "errors=%d\n", len(v.Validate()))
		return 0
	}
	if *sample == "scan" {
		fields := []formvalidate.Field{}
		rules := []formvalidate.Rule{}
		for index := 0; index < 50; index++ {
			name := fmt.Sprintf("f-%d", index)
			fields = append(fields, formvalidate.Field{Name: name})
			rules = append(rules, formvalidate.Rule{
				ID: fmt.Sprintf("r-%d", index), Field: name, Kind: "range", Min: 0, Max: 100,
			})
		}
		v, err := formvalidate.New(fields, rules)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		for _, field := range fields {
			v.Set(field.Name, "50")
		}
		for index := 0; index < 300; index++ {
			v.Validate()
		}
		fmt.Fprintf(stdout, "scanned=%d\n", v.Scanned())
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample required|bound|cross|order|parse|scan")
	return 2
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
