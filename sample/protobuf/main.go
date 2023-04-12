package main

import (
	"fmt"
	"log"

	"github.com/lxt1045/blog/sample/protobuf/gen"

	"github.com/golang/protobuf/proto"
)

func main() {
	// 为 AllPerson 填充数据
	p1 := gen.Person0{
		Id:   *proto.Int32(1),
		Name: *proto.String("hello world"),
	}

	p2 := gen.Person0{
		Id:   2,
		Name: "gopher",
	}

	all_p := gen.AllPerson{
		Per: []*gen.Person0{&p1, &p2},
	}

	// 对数据进行序列化
	data, err := proto.Marshal(&all_p)
	if err != nil {
		log.Fatalln("Mashal data error:", err)
	}

	// 对已经序列化的数据进行反序列化
	var target gen.AllPerson
	err = proto.Unmarshal(data, &target)
	if err != nil {
		log.Fatalln("UnMashal data error:", err)
	}

	fmt.Printf("%+v", target.Per) // 打印第一个 person Name 的值进行反序列化验证
}
