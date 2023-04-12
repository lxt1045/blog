package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/bytedance/sonic"
	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	"github.com/lxt1045/blog/sample/protobuf/gen"
	gogofastgen "github.com/lxt1045/blog/sample/protobuf/gogofastgen/gen"
	gogogen "github.com/lxt1045/blog/sample/protobuf/gogogen/gen"
	gogogen2 "github.com/lxt1045/blog/sample/protobuf/gogogen/gen2"
	personst "github.com/lxt1045/blog/sample/protobuf/person"
	lxtjson "github.com/lxt1045/json"
	"github.com/mailru/easyjson"
	"github.com/tinylib/msgp/msgp"
)

// https://www.cnblogs.com/zhangchaoyang/p/15256978.html

func BenchmarkEncode(b *testing.B) {
	b.Run("BenchmarkJsonEncode-std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(&doc)
		}
	})

	b.Run("BenchmarkJsonDecode-std", func(b *testing.B) {
		bs, _ := json.Marshal(&doc)
		var inst gen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkEasyJsonEncode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			person.MarshalJSON()
		}
	})

	b.Run("BenchmarkEasyJsonDecode", func(b *testing.B) {
		bs, _ := person.MarshalJSON()
		var inst personst.Person
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			easyjson.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkJsonEncode-sonic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.Marshal(&doc)
		}
	})

	b.Run("BenchmarkJsonDecode-sonic", func(b *testing.B) {
		bs, _ := json.Marshal(&doc)
		var inst gen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sonic.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkJsonEncode-lxt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lxtjson.Marshal(&doc)
		}
	})

	b.Run("BenchmarkJsonDecode-lxt", func(b *testing.B) {
		bs, _ := json.Marshal(&doc)
		var inst gen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			lxtjson.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkGobEncode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buffer bytes.Buffer
			encoder := gob.NewEncoder(&buffer)
			encoder.Encode(doc)
		}
	})

	b.Run("BenchmarkGobDecode", func(b *testing.B) {
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		encoder.Encode(doc)
		var inst gen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buffer.Reset()
			decoder := gob.NewDecoder(&buffer)
			decoder.Decode(&inst)
		}
	})

	b.Run("BenchmarkPbEncode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			proto.Marshal(&doc)
		}
	})

	b.Run("BenchmarkPbDecode", func(b *testing.B) {
		bs, _ := gogoproto.Marshal(&doc)
		var inst gen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			proto.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkPbEncode-gogo", func(b *testing.B) {
		bs, _ := gogoproto.Marshal(&doc)
		var doc gogogen.Doc
		gogoproto.Unmarshal(bs, &doc)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gogoproto.Marshal(&doc)
		}
	})

	b.Run("BenchmarkPbDecode-gogo", func(b *testing.B) {
		bs, _ := gogoproto.Marshal(&doc)
		var inst gogogen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gogoproto.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkPbEncode-gogofast", func(b *testing.B) {
		bs, _ := gogoproto.Marshal(&doc)
		var doc gogogen.Doc
		gogoproto.Unmarshal(bs, &doc)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gogoproto.Marshal(&doc)
		}
	})

	b.Run("BenchmarkPbDecode-gogofast", func(b *testing.B) {
		bs, _ := gogoproto.Marshal(&doc)
		var inst gogofastgen.Doc
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gogoproto.Unmarshal(bs, &inst)
		}
	})

	b.Run("BenchmarkMsgpEncode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			msgp.Encode(&buf, &person)
		}
	})
	b.Run("BenchmarkMsgpDecode", func(b *testing.B) {
		var buf bytes.Buffer
		msgp.Encode(&buf, &person)
		var inst personst.Person
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			msgp.Decode(&buf, &inst)
		}
	})
}

func BenchmarkGogo(b *testing.B) {
	list := []gogoproto.Message{
		&gogogen.Doc{},
		&gogogen2.Doc{},
	}
	for i, msg := range list {
		bs, _ := gogoproto.Marshal(&doc)
		gogoproto.Unmarshal(bs, msg)
		b.Run("Encode-gogo-"+strconv.Itoa(i), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				gogoproto.Marshal(msg)
			}
		})
		b.Run("Encode-gogo-"+strconv.Itoa(i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				gogoproto.Unmarshal(bs, msg)
			}
		})
	}
}
