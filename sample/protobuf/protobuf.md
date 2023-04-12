修饰符：

required : 　不可以增加或删除的字段，必须初始化；
optional : 　 可选字段，可删除，可以不初始化；
repeated : 　可重复字段， 对应到java文件里，生成的是List。

就会生成 xxx.pb.go 文件
protoc --go_out=./gen test.proto
protoc -I=. test.proto --gogofaster_out=./gogogen
protoc -I=. test.proto --gogofast_out=./gogofastgen

GO111MODULE="off" go get -u github.com/golang/protobuf/{protoc-gen-go,proto}
GO111MODULE="off" go get -u github.com/gogo/protobuf/protoc-gen-gogofaster

go install github.com/mailru/easyjson/...@latest
执行命令 easyjson -all ./person/person.go 会生成person_easyjson.go。

执行命令 go generate ./person/person.go 会生成person_gen.go和person_gen_test.go。