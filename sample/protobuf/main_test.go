package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"testing"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	"github.com/lxt1045/blog/sample/protobuf/gen"
	gogogen "github.com/lxt1045/blog/sample/protobuf/gogogen/gen"
	personst "github.com/lxt1045/blog/sample/protobuf/person"
	"github.com/mailru/easyjson"
	"github.com/tinylib/msgp/msgp"
)

var doc = gen.Doc{DocId: 123, Position: "搜索工程师", Company: "百度", City: "北京", SchoolLevel: 2, Vip: false, Chat: true, Active: 1, WorkAge: 3}
var person = personst.Person{DocId: 123, Position: "搜索工程师", Company: "百度", City: "北京", SchoolLevel: 2, Vip: false, Chat: true, Active: 1, WorkAge: 3}

func TestJson(t *testing.T) {
	bs, _ := json.Marshal(doc)
	fmt.Printf("json encode byte length %d\n", len(bs))
	var inst gen.Doc
	_ = json.Unmarshal(bs, &inst)
	fmt.Printf("json decode position %s\n", inst.Position)
}

func TestEasyJson(t *testing.T) {
	bs, _ := person.MarshalJSON()
	fmt.Printf("easyjson encode byte length %d\n", len(bs))
	var inst personst.Person
	_ = easyjson.Unmarshal(bs, &inst)
	fmt.Printf("easyjson decode position %s\n", inst.Position)
}

func TestGob(t *testing.T) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	_ = encoder.Encode(doc)
	fmt.Printf("gob encode byte length %d\n", len(buffer.Bytes()))
	var inst gen.Doc
	decoder := gob.NewDecoder(&buffer)
	_ = decoder.Decode(&inst)
	fmt.Printf("gob decode position %s\n", inst.Position)
}

func TestProtobuf(t *testing.T) {
	bs, _ := proto.Marshal(&doc)
	fmt.Printf("pb encode byte length %d\n", len(bs))
	var inst gen.Doc
	_ = proto.Unmarshal(bs, &inst)
	fmt.Printf("pb decode position %s\n", inst.Position)
}

func TestGogoProtobuf(t *testing.T) {
	bs, _ := gogoproto.Marshal(&doc)
	fmt.Printf("pb encode byte length %d\n", len(bs))
	var inst gogogen.Doc
	_ = gogoproto.Unmarshal(bs, &inst)
	fmt.Printf("pb decode position %s\n", inst.Position)
}

func TestMsgp(t *testing.T) {
	var buf bytes.Buffer
	_ = msgp.Encode(&buf, &person)
	fmt.Printf("msgp encode byte length %d\n", len(buf.Bytes()))
	var inst personst.Person
	_ = msgp.Decode(&buf, &inst)
	fmt.Printf("msgp decode position %s\n", inst.Position)
}
