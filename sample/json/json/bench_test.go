package json_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/bytedance/sonic"
	lxt "github.com/lxt1045/blog/sample/json/json"
	"github.com/tidwall/gjson"
)

func init() {
	fFromFile := func(data []string, name string) []string {
		bs, err := ioutil.ReadFile(name)
		if err != nil {
			panic(err)
		}
		data = append(data, BytesToString(bs))
		return data
	}
	data = fFromFile(data, "./testdata/twitter.json")
	data = fFromFile(data, "./testdata/twitterescaped.json")
}

func TestMyUnmarshal0(t *testing.T) {

	var m datastruct0
	bs := StringToBytes(data[0])
	lxt.Unmarshal(bs, &m)

	bs, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("to:%s", string(bs))
}

// go test -benchmem -run=^$ -bench "^(BenchmarkUnmarshal)$" github.com/lxt1045/blog/sample/json/json -count=1 -v -cpuprofile cpu.prof -c
// go tool pprof ./json.test cpu.prof
// web
func BenchmarkUnmarshal(b *testing.B) {
	runs := []struct {
		name string
		f    func(string)
	}{
		{"std",
			func(data string) {
				var m map[string]interface{}
				bs := StringToBytes(data)
				json.Unmarshal(bs, &m)
			},
		},
		{"gjson",
			func(data string) {
				gjson.Parse(data).Value()
			},
		},
		{
			"sonic",
			func(data string) {
				var m map[string]interface{}
				sonic.UnmarshalString(data, &m)
			},
		},
	}
	// runs = runs[2:]
	for _, d := range data {
		b.Logf("len(data):%.3fk\n", float64(len(d))/1000.0)
		for _, r := range runs {
			name := fmt.Sprintf("%d-%s", len(d), r.name)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					r.f(d)
				}
				b.SetBytes(int64(b.N))
				b.StopTimer()
			})
		}
	}
}
func BenchmarkUnmarshalStruct0(b *testing.B) {
	bs := StringToBytes(data[0])
	d := string(bs)
	var m datastruct0
	lxt.Unmarshal(bs, &m)
	runs := []struct {
		name string
		f    func()
	}{
		{"std",
			func() {
				var m map[string]interface{}
				json.Unmarshal(bs, &m)
			},
		},
		{"std-st",
			func() {
				var m datastruct0
				json.Unmarshal(bs, &m)
			},
		},
		{"lxt-st",
			func() {
				var m datastruct0
				lxt.Unmarshal(bs, &m)
			},
		},
		{
			"sonic",
			func() {
				var m map[string]interface{}
				sonic.UnmarshalString(d, &m)
			},
		},
		{
			"sonic-st",
			func() {
				var m datastruct0
				sonic.UnmarshalString(d, &m)
			},
		},
	}
	for _, r := range runs {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.SetBytes(int64(b.N))
			b.StopTimer()
		})
	}
}

func BenchmarkUnmarshalStruct01(b *testing.B) {
	bs := StringToBytes(data[1])
	d := string(bs)
	runs := []struct {
		name string
		f    func()
	}{
		{"std",
			func() {
				var m map[string]interface{}
				json.Unmarshal(bs, &m)
			},
		},
		{"std-st",
			func() {
				var m datastruct1
				json.Unmarshal(bs, &m)
			},
		},
		{"lxt-st",
			func() {
				var m datastruct1
				lxt.Unmarshal(bs, &m)
			},
		},
		{
			"sonic",
			func() {
				var m map[string]interface{}
				sonic.UnmarshalString(d, &m)
			},
		},
		{
			"sonic-st",
			func() {
				var m datastruct1
				sonic.UnmarshalString(d, &m)
			},
		},
	}
	var m datastruct1
	lxt.Unmarshal(bs, &m)
	// runs = runs[2:]
	for _, d := range data[1:2] {
		b.Logf("len(data):%.3fk\n", float64(len(d))/1000.0)
		for _, r := range runs {
			name := fmt.Sprintf("%d-%s", len(d), r.name)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					r.f()
				}
				b.SetBytes(int64(b.N))
				b.StopTimer()
			})
		}
	}
}

type datastruct1 struct {
	ItemID         int64       `json:"ItemID"`
	BizName        BizName     `json:"BizName"`
	BizCode        string      `json:"BizCode"`
	Description    Description `json:"Description"`
	Type           int         `json:"Type"`
	ItemManagerURL string      `json:"ItemManagerURL"`
	ItemEnumURL    string      `json:"ItemEnumURL"`
}
type datastruct0 struct {
	Apps []Apps `json:"Apps"`
}
type BizName struct {
	ZHCN string `json:"ZH_CN"`
	ENUS string `json:"EN_US"`
}
type Description struct {
	ZHCN string `json:"ZH_CN"`
	ENUS string `json:"EN_US"`
}
type ControlItem struct {
	ItemID         int64       `json:"ItemID"`
	BizName        BizName     `json:"BizName"`
	BizCode        string      `json:"BizCode"`
	Description    Description `json:"Description"`
	Type           int         `json:"Type"`
	ItemManagerURL string      `json:"ItemManagerURL"`
	ItemEnumURL    string      `json:"ItemEnumURL"`
}
type Managers struct {
	ID       int64  `json:"ID"`
	UserID   string `json:"UserID"`
	TenantID string `json:"TenantID"`
	FromApp  bool   `json:"FromApp"`
}
type Name struct {
	ZHCN string `json:"ZH_CN"`
	ENUS string `json:"EN_US"`
}
type ControlStrategy struct {
	StrategyID    int64       `json:"StrategyID"`
	GiveStrategy  int         `json:"GiveStrategy"`
	BuildStrategy int         `json:"BuildStrategy"`
	Name          Name        `json:"Name"`
	Description   Description `json:"Description"`
}
type EnumName struct {
	ZHCN string `json:"ZH_CN"`
	ENUS string `json:"EN_US"`
}
type ControlItemEnums struct {
	ID            int64    `json:"ID"`
	EnumName      EnumName `json:"EnumName"`
	EnumCode      string   `json:"EnumCode"`
	TenantID      string   `json:"TenantID"`
	ControlItemID int      `json:"ControlItemID"`
	Source        int      `json:"Source"`
	Status        int      `json:"Status"`
}
type BuControlItemEnums struct {
	ID            int64    `json:"ID"`
	EnumName      EnumName `json:"EnumName"`
	EnumCode      string   `json:"EnumCode"`
	TenantID      string   `json:"TenantID"`
	ControlItemID int      `json:"ControlItemID"`
	Source        int      `json:"Source"`
	Status        int      `json:"Status"`
}
type BuControlItems struct {
	ControlItem             ControlItem          `json:"ControlItem"`
	Managers                []Managers           `json:"Managers"`
	ControlStrategy         ControlStrategy      `json:"ControlStrategy"`
	ControlItemEnums        []ControlItemEnums   `json:"ControlItemEnums"`
	BuControlItemEnums      []BuControlItemEnums `json:"BuControlItemEnums"`
	ManagerType             int                  `json:"ManagerType"`
	ManagerLevel            int                  `json:"ManagerLevel"`
	ManagersWritePermission bool                 `json:"ManagersWritePermission"`
	StrategyWritePermission bool                 `json:"StrategyWritePermission"`
	EnumsWritePermission    bool                 `json:"EnumsWritePermission"`
	BuEnumsWritePermission  bool                 `json:"BuEnumsWritePermission"`
}
type Apps struct {
	AppID          string           `json:"AppID"`
	BuControlItems []BuControlItems `json:"BuControlItems"`
}

var data = []string{
	`{
		"Apps": [
			{
				"AppID": "6914653959508461065",
				"BuControlItems": [
					{
						"ControlItem": {
							"ItemID": 1442408958374608801,
							"BizName": {
								"ZH_CN": "职级",
								"EN_US": "职级"
							},
							"BizCode": "JOB_LEVEL",
							"Description": {
								"ZH_CN": "",
								"EN_US": ""
							},
							"Type": 1,
							"ItemManagerURL": "",
							"ItemEnumURL": ""
						},
						"Managers": [
							{
								"ID": 1457001660292890625,
								"UserID": "6963170731978917413",
								"TenantID": "",
								"FromApp": true
							},
							{
								"ID": 1456541347368443906,
								"UserID": "6970583941266818604",
								"TenantID": "",
								"FromApp": true
							}
						],
						"ControlStrategy": {
							"StrategyID": 1442408958374608823,
							"GiveStrategy": 2,
							"BuildStrategy": 1,
							"Name": {
								"ZH_CN": "",
								"EN_US": ""
							},
							"Description": {
								"ZH_CN": "部分上级分配,允许自建",
								"EN_US": ""
							}
						},
						"ControlItemEnums": [
							{
								"ID": 1442408958374608801,
								"EnumName": {
									"ZH_CN": "1-1",
									"EN_US": "1-1"
								},
								"EnumCode": "",
								"TenantID": "",
								"ControlItemID": 0,
								"Source": 0,
								"Status": 0
							},
							{
								"ID": 1442408958374608802,
								"EnumName": {
									"ZH_CN": "1-2",
									"EN_US": "1-2"
								},
								"EnumCode": "",
								"TenantID": "",
								"ControlItemID": 0,
								"Source": 0,
								"Status": 0
							}
						],
						"BuControlItemEnums": [
							{
								"ID": 1457606571718193152,
								"EnumName": {
									"ZH_CN": "8-8",
									"EN_US": "8-8"
								},
								"EnumCode": "",
								"TenantID": "",
								"ControlItemID": 0,
								"Source": 0,
								"Status": 0
							}
						],
						"ManagerType": 2,
						"ManagerLevel": 1,
						"ManagersWritePermission": false,
						"StrategyWritePermission": false,
						"EnumsWritePermission": false,
						"BuEnumsWritePermission": true
					}
				]
			}
		]
	}`,
	`{
		"ItemID": 1442408958374608801,
		"BizName": {
			"ZH_CN": "职级",
			"EN_US": "职级"
		},
		"BizCode": "JOB_LEVEL",
		"Description": {
			"ZH_CN": "",
			"EN_US": ""
		},
		"Type": 1,
		"ItemManagerURL": "",
		"ItemEnumURL": ""
	}`,
}

//BytesToString ...
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//StringToBytes ...
func StringToBytes(s string) []byte {
	strH := (*reflect.StringHeader)(unsafe.Pointer(&s))
	p := reflect.SliceHeader{
		Data: strH.Data,
		Len:  strH.Len,
		Cap:  strH.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&p))
}
