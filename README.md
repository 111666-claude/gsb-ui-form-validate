# ui-form-validate

设置面板校验：必填、范围、跨字段规则与错误排序，只用 Go 标准库。

```
go run ./cmd/ui-form-validate --sample=required
go run ./cmd/ui-form-validate --sample=bound
go run ./cmd/ui-form-validate --sample=cross
go run ./cmd/ui-form-validate --sample=order
go run ./cmd/ui-form-validate --sample=parse
go run ./cmd/ui-form-validate --sample=scan
go test ./...
go vet ./...
```

## 口径（README 为准）

- **必填**：`Required` 的字段去掉首尾空白后为空时报一条 `required` 错误；该字段本轮的其它规则不再判定。
- **范围**：`range` 规则要求数值落在 `[Min, Max]` 之间，正好等于端点算通过；值解析不成数值也算一条错误。
- **跨字段**：`le` 要求本字段数值不大于 `Other` 字段的数值，`ge` 要求不小于；`Other` 字段没有设过值
  或解析不成数值时本条规则跳过。
- **排序**：错误按字段在定义表里的顺序升序，同字段按规则 `Priority` 升序，再按规则 ID 升序。
- **索引**：规则按字段建索引，校验一个字段不许扫描全部规则，`Scanned` 单次不超过字段数加命中规则数。
- **不变量**：同一份表单重复校验结果逐条一致，校验不改变表单。
- **规模**：1 万个字段与规则、每秒 10 万次校验。

## 输出契约（不改格式）

```
errors=..
order=..
scanned=..
```

## 目录

```
validate.go           必填、范围、跨字段与排序
validate_test.go      单元测试
cmd/ui-form-validate/ 命令行入口
```
