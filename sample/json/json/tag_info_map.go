package json

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sort"

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
	_ = idxs
	// log.Printf("idxs:%+v;idxRet:%+v", idxs, m.idxN)
	nBit := len(m.idxN)
	m.N = getNLen(nBit)
	m.S = make([]mapNode, m.N)
	if len(m.idxN) == 0 {
		err := lxterrs.New("buildMap:%+v, idxRet:%+v", nodes, m.idxN)
		panic(err)
	}
	m.idxNTable = getHashParam(m.idxN)

	for _, n := range nodes {
		idx := hash(n.K, m.idxNTable, m.idxN)
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

type iN struct {
	iByte int // []byte的偏移量
	mask  byte
	iBit  int
}

type Mask struct {
	iByte int
	mask  byte
	diff  float32
	nMask uint32 // 命中 mask 的数量; (mask & b) == mask
}

func logicalHash2(bsList [][]byte) (idxN []iN, idxRet []int) {
	var idxsRet [][]int
	if len(bsList) <= 1 {
		idxN = []iN{{
			iByte: 0, // key 的偏移量
			mask:  1,
			iBit:  1,
		}}
		return
	}

	idxsRet = divide2(bsList)
	// log.Printf("---idxsRet---:%+v", idxsRet)
	// PrintKeys(bsList)
	idxRet = idxsRet[0]

	idxN = make([]iN, len(idxRet))
	for i, x := range idxRet {
		iByte := x / 8 // key 的偏移量
		iBit := x % 8  // byte 内部偏移量
		idxN[i] = iN{
			iByte: iByte, // key 的偏移量
			mask:  1 << (7 - iBit),
			iBit:  1 << i,
		}
	}
	sort.Slice(idxN, func(i, j int) bool { return idxN[i].iByte < idxN[j].iByte })
	return
}

//idxRet:[{iByte:1 mask:1 iBit:3} {iByte:2 mask:1 iBit:1} {iByte:3 mask:1 iBit:7} {iByte:4 mask:1 iBit:6}
//{iByte:5 mask:1 iBit:0} {iByte:6 mask:1 iBit:5} {iByte:8 mask:1 iBit:4} {iByte:12 mask:1 iBit:2}]
// "name": "avatar",
// 01234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567891012345678911123456789121234567
// 00100010 01101110 01100001 01101101 01100101 00100010:"name"
// 00100010 01100001 01110110 01100001 01110100 011000010111001000100010:"avatar"
func logicalHash(bsList [][]byte) (idxN []iN, idxRet []int) {
	if len(bsList) <= 1 {
		idxN = []iN{{
			iByte: 0, // key 的偏移量
			mask:  1,
			iBit:  1,
		}}
		return
	}

	masks := getPivotMask(bsList)
	for i, m := range masks {
		idxN = append(idxN, iN{
			iByte: m.iByte,
			mask:  m.mask,
			iBit:  1 << i,
		})
	}
	// log.Printf("---idxN---:%+v", idxN)
	sort.Slice(idxN, func(i, j int) bool { return idxN[i].iByte < idxN[j].iByte })
	// for i := range idxN {
	// 	idxN[i].iBit = 1 << i
	// }
	return
}

// 统一分块式处理： 之前是n块，之后分的块越多就表示越好，即就越应该选该 bit 作为下一 bit
// 通过染色方式来表达是否已分区，比如 1 3 5 给打个 x 标签，当前有 n 个标签（即 n 个区），下一次 13 的标签变成 y，则就变成了 n+1 个区块了；
// 简单的按区来 进行两层遍历就好了！！！ 完美

func getPivotMask(bsList [][]byte) (ms []Mask) {
	lMax := len(bsList[0])
	for _, bs := range bsList {
		if lMax < len(bs) {
			lMax = len(bs)
		}
	}

	type Block struct {
		iByte     int
		NO        int // 区块编号
		maskCount int // 命中掩码的数量
	}
	type Hit struct {
		hit         bool
		blockNO     int
		nextBlockNO int // 下一级的；延续出来的
		nMask       float32
		nNextMask   float32
	}
	type BlockNew struct {
		mask  byte
		iByte int
		N     int
		list  []Block

		// 评价体系
		newBlock int
		diffSum  float32 // 所有区块 距离中线的分数的和
	}
	blocks := []BlockNew{}
	nextBlockNO := 1
	fNextBlock := func() (NO int) {
		NO = nextBlockNO
		nextBlockNO++
		return
	}

	// 设计提前退出的策略
start:
	blockNewMax := BlockNew{diffSum: math.MaxInt}
outfor:
	for iByte := 0; iByte < lMax; iByte++ {
		for _, mask := range allMask {
			block := BlockNew{
				list: make([]Block, len(bsList)),
			}
			if len(blocks) > 0 {
				copy(block.list, blocks[len(blocks)-1].list)
			}

			mBlockHit := make(map[int]*Hit) // 取款已命中 block
			newBlock := 0                   // 此次循环新增的区块

			for i, bs := range bsList {
				hit := false
				if len(bs) > iByte {
					hit = (mask & bs[iByte]) == mask // 是否命中掩码 mask
				}
				blockNO := block.list[i].NO // 当前 bs 的区块号码
				blockHit := mBlockHit[blockNO]
				if blockHit == nil {
					mBlockHit[blockNO] = &Hit{
						hit:     hit,
						blockNO: blockNO,
						nMask:   1,
					}
					continue
				}
				if blockHit.hit == hit {
					blockHit.nMask++
					continue
				}
				if blockHit.nextBlockNO == 0 {
					newBlock++
					blockHit.nextBlockNO = fNextBlock()
				}
				blockHit.nNextMask++
				block.list[i].NO = blockHit.nextBlockNO
				block.list[i].iByte = iByte
			}
			if newBlock == 0 {
				continue
			}

			var diffSum float32 = 0
			for _, hit := range mBlockHit {
				if hit.nMask > 1 || hit.nNextMask > 1 {
					diff := hit.nMask - hit.nNextMask
					diff = diff * diff
					if diff == 1 {
						diff = 0
					}
					//diff 要用 nMask 做一下修正？
					diff = diff / (hit.nMask + hit.nNextMask)
					// diff = diff / float32(newBlock+1)
					diffSum += diff
				}
			}
			// diffSum = diffSum / float32(newBlock*newBlock)

			// if blockNewMax.newBlock < newBlock || blockNewMax.diffSum > diffSum {
			if blockNewMax.diffSum > diffSum {
				blockNewMax = BlockNew{
					newBlock: newBlock,
					iByte:    iByte,
					mask:     mask,
					list:     block.list,
					diffSum:  diffSum,
				}
			}
			m := map[int]bool{}
			for _, b := range block.list {
				if !m[b.NO] {
					block.N++
					m[b.NO] = true
				}
			}
			if block.N == len(bsList) {
				ms = append(ms, Mask{iByte: iByte, mask: mask, diff: float32(blockNewMax.diffSum)})
				blocks = append(blocks, block)
				// log.Printf("len:%d, blocks:%+v", len(blocks), blocks)
				return
			}
			if blockNewMax.diffSum == 0 {
				// blocks = append(blocks, blockNewMax)
				// ms = append(ms, Mask{iByte: blockNewMax.iByte, mask: blockNewMax.mask})
				// goto start
				break outfor
			}
		}
	}
	blocks = append(blocks, blockNewMax)
	ms = append(ms, Mask{iByte: blockNewMax.iByte, mask: blockNewMax.mask, diff: float32(blockNewMax.diffSum)})
	goto start
}

func getPivot(bsList [][]byte) (min Mask) {
	lMax := len(bsList[0])
	for _, bs := range bsList {
		if lMax < len(bs) {
			lMax = len(bs)
		}
	}
	l := uint32(len(bsList))
	midL := float32(l) / 2 //获取中值

	// 0b0001, 0b0010, 0b0011, 0b0100 ...
	allMask := make([]byte, 0, 0xff)
	for i := 1; i < 0xff; i++ {
		allMask = append(allMask)
	}
	for _, mask := range allMask {
		masks := make([]Mask, lMax) // bit位为 1 的个数
		for _, bs := range bsList {
			for i, b := range bs {
				// masks[i].iByte = i
				if (mask & b) == mask {
					masks[i].nMask++
				}
			}
		}

		for i, mm := range masks {
			if mm.nMask == 0 || mm.nMask == l {
				continue
			}

			diff := midL - float32(mm.nMask)
			if diff < 0 {
				diff = -diff
			}
			if min.nMask == 0 || min.diff > diff {
				min = Mask{
					iByte: i,
					diff:  diff,
					nMask: mm.nMask,
					mask:  mask,
				}
				// 如果已经均分,则提前返回
				if min.diff < 0.51 {
					return min
				}
			}
		}
	}
	return min
}

func getHashParam(idxN []iN) (idxNTable []int16) {
	// 可以提前聚合成 byte ，然后打表合并 bit 增速最高 8 倍？
	idxNTable = make([]int16, 1024) //idxN[len(idxN)-1].iByte+1)
	j := int16(0)
	for i := range idxNTable {
		if int(j) == len(idxN) || idxN[j].iByte >= i {
			idxNTable[i] = j
			continue
		}
		for int(j) < len(idxN) {
			if int(j) == len(idxN) || idxN[j].iByte >= i {
				break
			}
			j++
		}
		idxNTable[i] = j
	}
	return
}

func hash2(key []byte, idxNTable []int16, idxN []iN) (idx int) {
	n := idxNTable[len(key)]
	for _, x := range idxN[:n] {
		if (x.mask & key[x.iByte]) == x.mask {
			idx |= x.iBit
		}
		// TODO
		/*
			可以改进，用 uint64 来匹配：
			flag := 0b0010010010
			if (in & flag) ^ flag == 0 {
				// 表示匹配
			}
		*/
	}
	return
}

func hash(key []byte, idxNTable []int16, idxN []iN) (idx int) {
	for _, x := range idxN {
		if x.iByte >= len(key) {
			break
		}
		if (x.mask & key[x.iByte]) == x.mask {
			idx |= x.iBit
		}
	}
	return
}
