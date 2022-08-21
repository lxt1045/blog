module github.com/lxt1045/blog/sample/json/json

go 1.18

require (
	github.com/bytedance/sonic v1.3.5
	github.com/lxt1045/Experiment/golang/json/pkg/json v0.0.0-00010101000000-000000000000
	github.com/lxt1045/errors v0.0.0-20211214155050-af6c6c19b840
	github.com/tidwall/gjson v1.13.0
	github.com/tidwall/match v1.1.1
	github.com/tidwall/pretty v1.2.0
)

require (
	github.com/chenzhuoyu/base64x v0.0.0-20211019084208-fb5309c8db06 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
)

replace (
	github.com/lxt1045/Experiment/golang/json/pkg/json => /Users/bytedance/go/src/github.com/lxt1045/Experiment/golang/json/pkg/json
	github.com/lxt1045/errors => /Users/bytedance/go/src/github.com/lxt1045/errors
)
