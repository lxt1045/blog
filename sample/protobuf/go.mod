module github.com/lxt1045/blog/sample/protobuf

go 1.18

require (
	github.com/bytedance/sonic v1.8.2
	github.com/gogo/protobuf v1.3.3-0.20221024144010-f67b8970b736
	github.com/golang/protobuf v1.5.3
	github.com/lxt1045/json v0.0.0-20230406161715-3fd66395f845
	github.com/mailru/easyjson v0.7.7
	github.com/tinylib/msgp v1.1.8
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/lxt1045/errors v0.0.0-20211214155050-af6c6c19b840 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
)

replace (
	github.com/gogo/protobuf => github.com/lxt1045/protobuf v0.0.0-20221024144010-f67b8970b736
	github.com/lxt1045/errors => /Users/bytedance/go/src/github.com/lxt1045/errors
	github.com/lxt1045/json => /Users/bytedance/go/src/github.com/lxt1045/json
)
