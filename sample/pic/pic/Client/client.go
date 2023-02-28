package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile) //log.Llongfile
}
func httpGET(host, url string) {
	resp, err := http.Get("http://" + host + url)
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	fmt.Println(string(body))
}

func httpGet(path string, header map[string]string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if nil != err {
		return
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	return
}

func httpDo(url, path, method, body string) (req *http.Request, resp *http.Response, err error) {
	if method == http.MethodGet {
		req, err = http.NewRequest(http.MethodGet, "http://"+url+path, nil)
	} else {
		buf := bytes.NewBuffer([]byte(body))
		req, err = http.NewRequest(method, "http://"+url+path, buf)
	}
	if nil != err {
		return
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	return
}
func main() {
	url := "https://image.baidu.com/search/acjson?tn=resultjson_com&logid=9306699209664286820&ipn=rj&ct=201326592&is=&fp=result&fr=&word=雷军&queryWord=雷军&cl=2&lm=-1&ie=utf-8&oe=utf-8&adpicid=&st=&z=&ic=&hd=&latest=&copyright=&s=&se=&tab=&width=&height=&face=&istype=&qc=&nc=&expermode=&nojc=&isAsync=&pn=300&rn=30&gsm=12c&1676201710852="
	//httpGET("127.0.0.1:8210", "/")
	resp, err := httpGet(url, map[string]string{
		"Cookie":  `BDqhfp=雷军&&NaN-1undefined&&7140&&9; __yjs_duid=1_e4e57be84fa2de1cb88d42d68185a8371629870243863; BIDUPSID=DD01E73F27ABB032C4C8FDE2A7C578A2; PSTM=1632303619; delPer=0; BDRCVFR[feWj1Vr5u3D]=I67x6TjHwwYf0; H_WISE_SIDS=107317_110085_127969_179348_184716_188748_189660_190618_191067_191252_192387_194085_194511_194529_195328_195342_195467_196050_196427_196514_196528_197241_197471_197711_197783_197833_197955_198267_198930_199082_199466_199489_199569_200150_200349_200487_200763_200958_201054_201191_201233_201553_201576_201979_202652_202759_202822_202832_202916_203197_203247_203257_203310_203520_203525_203606_203999_204099_204112_204132_204170_204260_204370_204431_204662_204855_204910_204940; ZD_ENTRY=google; BDRCVFR[tox4WRQ4-Km]=mk3SLVN4HKm; BDRCVFR[-pGxjrCMryR]=mk3SLVN4HKm; BDRCVFR[CLK3Lyfkr9D]=mk3SLVN4HKm; BAIDUID=A1D47FA385F156AC604EE949505645D0:FG=1; BAIDU_WISE_UID=wapp_1658402494391_222; ZFY=UYTRq3l19wV7neZ:B6NjPcGWUX00CvzYo6JqzTvxT:AtU:C; BDUSS=1llRlE1dnF2RmJLWDFzMnJBZ21QbTFuSXJWSzRBOHItQlZVVEpOWGo2eFR3d2hqRVFBQUFBJCQAAAAAAAAAAAEAAABvy0oa1sHT2s7Su7nKx7K70MUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFM24WJTNuFiR; BDUSS_BFESS=1llRlE1dnF2RmJLWDFzMnJBZ21QbTFuSXJWSzRBOHItQlZVVEpOWGo2eFR3d2hqRVFBQUFBJCQAAAAAAAAAAAEAAABvy0oa1sHT2s7Su7nKx7K70MUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFM24WJTNuFiR; MCITY=-:; BAIDUID_BFESS=A1D47FA385F156AC604EE949505645D0:FG=1; H_WISE_SIDS_BFESS=107317_110085_127969_179348_184716_188748_189660_190618_191067_191252_192387_194085_194511_194529_195328_195342_195467_196050_196427_196514_196528_197241_197471_197711_197783_197833_197955_198267_198930_199082_199466_199489_199569_200150_200349_200487_200763_200958_201054_201191_201233_201553_201576_201979_202652_202759_202822_202832_202916_203197_203247_203257_203310_203520_203525_203606_203999_204099_204112_204132_204170_204260_204370_204431_204662_204855_204910_204940; BA_HECTOR=8l0l2lag8ga52g812haha1g81huh50m1l; BDORZ=B490B5EBF6F3CD402E515D22BCDA1598; arialoadData=false; BDRCVFR[dG2JNJb_ajR]=mk3SLVN4HKm; indexPageSugList=["张一鸣"]; cleanHistoryStatus=0; PSINO=1; H_PS_PSSID=36549_38130_37910_37989_36803_37932_38089_38041_26350_37285_38098_38008_37881; RT="z=1&dm=baidu.com&si=2b82cef4-7573-4e56-8005-2f4754e2c9dd&ss=le19dm8w&sl=2&tt=7s7&bcn=https://fclog.baidu.com/log/weirwood?type=perf&nu=o086j0z7&cl=5z1&ul=5zd&ld=69x&hd=6kf"; userFrom=null; ab_sr=1.0.1_MTM0ZThjOTU3YTg0NTJhNWY3Nzc1NTE1MWY5MzBlMWZkNDAxODIwODY1YzcwMjc3MzI2MWVmNjNiNmFlM2NkMzY2MGZjNmQ2ZDdiY2YyOGIwN2I2ZmFmNGEzMDAyNjVjMGIxZTUyMGQzY2ZmNTA5NmM5ZTdkMjM3NzA2OTg5ZmUxM2VmNjAzNzQ1MzlhMTlkNzJlMjQ0Mzg1OWQwZTU3Nw==`,
		"Referer": "https://image.baidu.com/search/index?tn=baiduimage&ct=201326592&lm=-1&cl=2&dyTabStr=MCwxLDMsNiw0LDUsNyw4LDIsOQ==&word=雷军",
	})
	if err != nil {
		log.Printf("err:%v", err)
		return
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read err:%v", err)
		return
	}
	log.Printf("resp:%+v", string(bs))
}
