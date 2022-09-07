package json

import (
	"bytes"
	"fmt"
	"log"

	lxterrs "github.com/lxt1045/errors"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile)
}

type tagMap struct {
	S []mapNode
	N int

	idxNTable []int16
	idxN      []iN
}
type mapNode struct {
	K []byte
	V *TagInfo
}

func (m *tagMap) GetV(k []byte) (v *TagInfo) {
	idx := hash2(k, m.idxNTable, m.idxN)
	n := m.S[idx]
	if bytes.Equal(k, n.K) {
		return n.V
	}
	return
}
func (m *tagMap) Get(k []byte) (v *TagInfo) {
	idx := hash(k, nil, m.idxN)
	n := m.S[idx]
	if bytes.Equal(k, n.K) {
		return n.V
	}
	return
}
func (m *tagMap) Get2(k []byte) (v *TagInfo) {
	idx := hash2(k, m.idxNTable, m.idxN)
	n := m.S[idx]
	if bytes.Equal(k, n.K) {
		return n.V
	}
	return
}

func (m *tagMap) String() (str string) {
	str += fmt.Sprintf("len:%d, idxN:%+v;\nkeys:", len(m.S), m.idxN)
	keys := []string{}
	for i, n := range m.S {
		if len(n.K) == 0 {
			continue
		}
		keys = append(keys, string(n.K))
		str += fmt.Sprintf("[%d] %s: %+v;", i, n.K, n.V.TagName)
	}
	str += "\nbsList:\n"
	for i := 0; i < 128; i++ {
		if i%10 == 0 {
			str += fmt.Sprintf("%d", i/10)
			continue
		}
		str += fmt.Sprintf("%d", i%10)
	}
	for _, bs := range keys {
		str += "\n"
		for _, b := range bs {
			str += fmt.Sprintf("%08b", b)
		}
		str += fmt.Sprintf(":%s", string(bs))
	}
	return
}
func PrintKeys(bsList [][]byte) {
	str := "\nbsList:\n"
	for i := 0; i < 128; i++ {
		if i%10 == 0 {
			str += fmt.Sprintf("%d", i/10)
			continue
		}
		str += fmt.Sprintf("%d", i%10)
	}
	for _, bs := range bsList {
		str += "\n"
		for _, b := range bs {
			str += fmt.Sprintf("%08b", b)
		}
		str += fmt.Sprintf(":%s", string(bs))
	}
	log.Printf("%s\n", str)
}

func PrintKeys2(bsList [][]byte) {
	str := "\nbsList:\n"
	for i := 0; i < 32; i++ {
		str += fmt.Sprintf("%d", i%10)
	}
	log.Printf("%s", str)
	for _, bs := range bsList {
		for _, b := range bs {
			str += fmt.Sprintf("%08b", b)
		}
		log.Printf("%s:%s", str, string(bs))
	}
}

//找到醉倒的区分度(第 n 位)，二分
// 1bit 不行就 2bit
// 贪婪算法？全遍历？
// key 带上最后一个 " 以便于比较（第一个也可以带上）
func buildTagMap(nodes []mapNode) (m tagMap) {
	bsList := make([][]byte, 0, len(nodes))
	for _, n := range nodes {
		bsList = append(bsList, n.K)
	}
	// log.Printf("\n\n\n")
	// PrintKeys(bsList)
	// log.Printf("\n\n\n")
	idxs := []int{}
	m.idxN, idxs = logicalHash(bsList)
	log.Printf("idxs:%+v;idxRet:%+v", idxs, m.idxN)
	nBit := len(m.idxN)
	m.N = getNLen(nBit)
	m.S = make([]mapNode, m.N)
	if len(m.idxN) == 0 {
		err := lxterrs.New("buildMap:%+v, idxRet:%+v", nodes, m.idxN)
		panic(err)
	}
	m.idxNTable = getHashParam(m.idxN)

	for _, n := range nodes {
		idx := hash(n.K, nil, m.idxN)
		if nn := m.S[idx]; len(nn.K) > 0 && !bytes.Equal(nn.K, n.K) {
			PrintKeys(bsList)
			log.Printf("tagMap:%s", m.String())
			err := lxterrs.New("buildTagMap: key collision; %s: %s, idxRet:%+v",
				string(nn.K), string(n.K), m.idxN)
			panic(err)
		}

		m.S[idx] = n
	}
	// PrintKeys(bsList)
	// log.Printf("nBit:%d,len:%d,buildTagMap:\n%s", nBit, m.N, m.String())
	return
}
