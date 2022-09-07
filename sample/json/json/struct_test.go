package json

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	asrt "github.com/stretchr/testify/assert"
)

func TestStruct(t *testing.T) {
	type Anonymous struct {
		Count int `json:"count"`
		X     string
	}
	fLine := func() string {
		_, file, line, _ := runtime.Caller(1)
		_, file = filepath.Split(file)
		return file + ":" + strconv.Itoa(line)
	}
	idx := -3

	datas := []struct {
		name   string
		bs     string
		target string
		data   interface{}
	}{

		{
			name:   "interface:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int         `json:"out"`
				Struct interface{} `json:"struct_0"`
			}{},
		},
		{
			name:   "map" + fLine(),
			bs:     `{"out": 11 , "map_0": { "count":8,"y":"yyy"}}`,
			target: `{"out":11,"map_0":{"count":8,"y":"yyy"}}`,
			data:   &map[string]interface{}{},
		},

		// 匿名类型; 指针匿名类型
		{
			name:   "struct-Anonymous:" + fLine(),
			bs:     `{"out": 11 , "count":8,"X":"xxx"}`,
			target: `{"out":11,"count":8,"X":"xxx"}`,
			data: &struct {
				Out int `json:"out"`
				Anonymous
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int `json:"count"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8,"slice":[1,2,3]}}`,
			target: `{"out":11,"struct_0":{"count":8,"slice":[1,2,3]}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int   `json:"count"`
					Slice []int `json:"slice"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "slice:" + fLine(),
			bs:     `{"count":8 , "slice":[1,2,3] }`,
			target: `{"count":8,"slice":[1,2,3]}`,
			data: &struct {
				Count int   `json:"count"`
				Slice []int `json:"slice"`
			}{},
		},
		{
			name:   "bool:" + fLine(),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count bool `json:"count"`
				False bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr:" + fLine(),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr-null:" + fLine(),
			bs:     `{"count":true , "false_0":null }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr-empty:" + fLine(),
			bs:     `{"count":true }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "float64:" + fLine(),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count float64 `json:"count"`
			}{},
		},
		{
			name:   "float64-ptr:" + fLine(),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count *float64 `json:"count"`
			}{},
		},
		{
			name:   "int-ptr:" + fLine(),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count *int `json:"count"`
			}{},
		},
		{
			name:   "int:" + fLine(),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count int `json:"count"`
			}{},
		},
		{
			name:   "string-ptr:" + fLine(),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN *string
			}{},
		},
		{
			name:   "string-notag:" + fLine(),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN string
			}{},
		},
		{
			name:   "string:" + fLine(),
			bs:     `{ "ZH_CN":"chinese", "ENUS":"English", "count":8 }`,
			target: `{"ZH_CN":"chinese"}`,
			data: &struct {
				ZHCN string `json:"ZH_CN"`
			}{},
		},
	}
	if idx >= 0 {
		datas = datas[idx : idx+1]
	}

	for i, d := range datas {
		t.Run(d.name, func(t *testing.T) {
			err := Unmarshal([]byte(d.bs), d.data)
			if err != nil {
				t.Fatalf("[%d]%s, error:%v\n", i, d.name, err)
			}
			bs, err := json.Marshal(d.data)
			if err != nil {
				t.Fatalf("%s:%v\n", d.name, err)
			}
			if _, ok := (d.data).(*map[string]interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else if _, ok := (d.data).(*interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else {
				asrt.Equalf(t, d.target, string(bs), d.name)
			}
		})
	}
}
