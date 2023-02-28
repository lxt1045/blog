package hashmap

import (
	"fmt"
	"log"
	"math"
	"sort"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

//根据状态数 返回 bit 数
func getNLen(nBit int) (nStatus int) {
	nStatus = 1
	for i := 0; i < nBit; i++ {
		nStatus *= 2
	}
	return
}

var allMask = func() (allMask []byte) {
	allMask = make([]byte, 0, 256)
	seed := []byte{1, 2, 4, 8, 16, 32, 64, 128}

	// 1bit 的
	allMask = append(allMask, seed...)

	return
	// 2. 2bit 的
	for i, a := range seed {
		for _, b := range seed[i+1:] {
			allMask = append(allMask, a+b)
		}
	}

	// 3bit 的
	for i, a := range seed {
		for j, b := range seed[i+1:] {
			for _, c := range seed[i+j+2:] {
				allMask = append(allMask, a+b+c)
			}
		}
	}

	// 4bit 的
	for i, a := range seed {
		for j, b := range seed[i+1:] {
			for k, c := range seed[i+j+2:] {
				for _, d := range seed[i+j+k+3:] {
					allMask = append(allMask, a+b+c+d)
				}
			}
		}
	}
	// return
	// 5bit 的
	for i, a := range seed {
		for j, b := range seed[i+1:] {
			for k, c := range seed[i+j+2:] {
				for l, d := range seed[i+j+k+3:] {
					for _, e := range seed[i+j+k+l+4:] {
						allMask = append(allMask, a+b+c+d+e)
					}
				}
			}
		}
	}

	// 6bit 的
	for i, a := range seed {
		for j, b := range seed[i+1:] {
			for k, c := range seed[i+j+2:] {
				for l, d := range seed[i+j+k+3:] {
					for m, e := range seed[i+j+k+l+4:] {
						for _, f := range seed[i+j+k+l+m+5:] {
							allMask = append(allMask, a+b+c+d+e+f)
						}
					}
				}
			}
		}
	}

	// 7bit 的
	allMask = append(allMask, []byte{
		0b11111110, 0b11111101, 0b11111011, 0b11110111, 0b11101111,
		0b11011111, 0b10111111, 0b01111111, 0b11111111,
	}...)

	m := make(map[byte]struct{})
	for _, b := range allMask {
		if _, ok := m[b]; ok {
			panic(fmt.Sprintf("%b already exist", b))
		}
		m[b] = struct{}{}
	}

	for i := 1; i < 0xff; i++ {
		b := byte(i)
		if _, ok := m[b]; !ok {
			allMask = append(allMask, b)
			panic(fmt.Sprintf("%b not exist", b))
		}
	}
	return
}()

func PrintMask(allMask string) {
	str := "PrintMask:\n"
	for _, b := range allMask {
		str += fmt.Sprintf("%3d:%08b\n", b, b)
	}
	log.Printf("allMask:\n%s", str)
}

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile)
}

type tagMap struct {
	S         []mapNode
	idxNTable []int16
	idxN      []iN
	idxN2     []iN2
	N         int
}
type mapNode struct {
	K string
	V *TagInfo
}

func TagMapGetV(m *tagMap, k string) (v *TagInfo) {
	idx := hash2(k, m.idxNTable, m.idxN)
	n := m.S[idx]
	if k == n.K {
		return n.V
	}
	return
}
func TagMapGetV4(m *tagMap, k string) (v *TagInfo) {
	idx := hash4(k, m.idxNTable, m.idxN2)
	n := m.S[idx]
	if k == n.K {
		return n.V
	}
	return
}

func (m *tagMap) GetV(k string) (v *TagInfo) {
	idx := hash2(k, m.idxNTable, m.idxN)
	n := m.S[idx]
	if k == n.K {
		return n.V
	}
	return
}
func (m *tagMap) Get(k string) (v *TagInfo) {
	idx := hash(k, m.idxN)
	n := m.S[idx]
	if k == n.K {
		return n.V
	}
	return
}
func (m *tagMap) Get2(k string) (v *TagInfo) {
	idx := hash2(k, m.idxNTable, m.idxN)
	n := m.S[idx]
	if k == n.K {
		return n.V
	}
	return
}
func (m *tagMap) Get4(k string) (v *TagInfo) {
	idx := hash4(k, m.idxNTable, m.idxN2)
	n := m.S[idx]
	if k == n.K {
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
func PrintKeys(bsList []string) {
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

//找到醉倒的区分度(第 n 位)，二分
// 1bit 不行就 2bit
// 贪婪算法？全遍历？
// key 带上最后一个 " 以便于比较（第一个也可以带上）
func buildTagMap(nodes []mapNode) (m tagMap) {
	bsList := make([]string, 0, len(nodes))
	for _, n := range nodes {
		bsList = append(bsList, n.K)
	}
	// log.Printf("\n\n\n")
	// PrintKeys(bsList)
	// log.Printf("\n\n\n")
	idxs := []int{}
	m.idxN, idxs = logicalHash(bsList)
	for _, idx := range m.idxN {
		m.idxN2 = append(m.idxN2, iN2{
			iByte: int16(idx.iByte), // string的偏移量
			i:     int16(idx.i),
			iBit:  int16(idx.iBit),
			mask:  idx.mask,
		})

	}
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
		// idx := hash(n.K, m.idxNTable, m.idxN)
		idx := hash(n.K, m.idxN)
		if nn := m.S[idx]; len(nn.K) > 0 && nn.K != n.K {
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
	iByte   int // string的偏移量
	mask    byte
	iBit    int
	i       int
	ratio   float32 //占比
	diffSum float32 //分数
}

type iN2 struct {
	mask  byte
	iByte int16 // string的偏移量
	i     int16
	iBit  int16
}

type Mask struct {
	iByte   int
	mask    byte
	diffSum float32
	ratio   float32
	nMask   uint32 // 命中 mask 的数量; (mask & b) == mask
}

//idxRet:[{iByte:1 mask:1 iBit:3} {iByte:2 mask:1 iBit:1} {iByte:3 mask:1 iBit:7} {iByte:4 mask:1 iBit:6}
//{iByte:5 mask:1 iBit:0} {iByte:6 mask:1 iBit:5} {iByte:8 mask:1 iBit:4} {iByte:12 mask:1 iBit:2}]
// "name": "avatar",
// 01234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567891012345678911123456789121234567
// 00100010 01101110 01100001 01101101 01100101 00100010:"name"
// 00100010 01100001 01110110 01100001 01110100 011000010111001000100010:"avatar"
func logicalHash(bsList []string) (idxN []iN, idxRet []int) {
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
			iByte:   m.iByte,
			mask:    m.mask,
			iBit:    1 << i,
			i:       i,
			ratio:   m.ratio,
			diffSum: m.diffSum,
		})
	}
	// log.Printf("---idxN---:%+v", idxN)
	sort.Slice(idxN, func(i, j int) bool { return idxN[i].iByte < idxN[j].iByte })
	return
}

// 统一分块式处理： 之前是n块，之后分的块越多就表示越好，即就越应该选该 bit 作为下一 bit
// 通过染色方式来表达是否已分区，比如 1 3 5 给打个 x 标签，当前有 n 个标签（即 n 个区），下一次 13 的标签变成 y，则就变成了 n+1 个区块了；
// 简单的按区来 进行两层遍历就好了！！！ 完美

func getPivotMask1(bsList []string) (ms []Mask) {
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
		nextBlockNO int     // 下一级的；延续出来的
		nMask       float32 // 留在当前区块数量
		nNextMask   float32 // 分到新分裂区块的数量
	}
	type BlockNew struct {
		mask  byte
		iByte int
		N     int
		list  []Block

		// 评价体系
		newBlock int
		diffSum  float32 // 所有区块 距离中线的分数的和
		ratio    float32
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
				list: make([]Block, len(bsList)), // escapes to heap
			}
			if len(blocks) > 0 {
				copy(block.list, blocks[len(blocks)-1].list) // 继承最新一次区分好的列表
			}

			mBlockHit := make(map[int]*Hit) // 区块已命中 block
			newBlock := 0                   // 此次循环新增的区块

			for i, bs := range bsList {
				hit := false
				if len(bs) > iByte {
					hit = (mask & bs[iByte]) == mask // 是否命中掩码 mask
				}
				blockNO := block.list[i].NO // 当前 bs 的区块号码
				blockHit := mBlockHit[blockNO]
				if blockHit == nil {
					mBlockHit[blockNO] = &Hit{ // escapes to heap
						hit:     hit,
						blockNO: blockNO,
						nMask:   1,
					}
					continue
				}
				if blockHit.hit == hit {
					blockHit.nMask++ // 留在当前区块数量
					continue
				}
				if blockHit.nextBlockNO == 0 {
					newBlock++
					blockHit.nextBlockNO = fNextBlock()
				}
				blockHit.nNextMask++                    // 分到新分裂区块的数量
				block.list[i].NO = blockHit.nextBlockNO //
				block.list[i].iByte = iByte
			}
			if newBlock == 0 {
				continue
			}

			// diffSum==0 的情形都保留下来作为后备选项？
			var diffSum float32 = 0
			for _, hit := range mBlockHit {
				// 计算 欧拉距离
				diff := hit.nMask - hit.nNextMask
				if diff == 0 || diff == 1 || diff == -1 {
					continue
				}
				diff = diff / float32(newBlock*newBlock)
				diff = diff * diff
				//diff 要用 nMask 做一下修正？
				diff = diff / (hit.nMask + hit.nNextMask)
				// diff = diff / float32(newBlock*newBlock)
				diffSum += diff
			}
			// diffSum = diffSum / float32(newBlock*newBlock)
			// diffSum = float32(diffSum) / float32(newBlock*newBlock*newBlock)

			m := map[int]bool{}
			for _, b := range block.list {
				if !m[b.NO] {
					block.N++
					m[b.NO] = true
				}
			}
			ratio := float32(block.N) / float32(len(bsList))
			// if blockNewMax.newBlock < newBlock || blockNewMax.diffSum > diffSum {
			if blockNewMax.diffSum > diffSum {
				blockNewMax = BlockNew{
					newBlock: newBlock,
					iByte:    iByte,
					mask:     mask,
					list:     block.list,
					diffSum:  diffSum,
					ratio:    ratio,
				}
			}

			// 当前区块的数量和 bsList 一样，则表示达到结束条件
			if block.N == len(bsList) {
				ms = append(ms, Mask{iByte: iByte, mask: mask, diffSum: float32(blockNewMax.diffSum), ratio: blockNewMax.ratio})
				blocks = append(blocks, block) //
				// log.Printf("len:%d, blocks:%+v", len(blocks), blocks)
				return
			}
			if diffSum == 0 {
				// blocks = append(blocks, blockNewMax)
				// ms = append(ms, Mask{iByte: blockNewMax.iByte, mask: blockNewMax.mask})
				// goto start
				break outfor
			}
		}
	}
	blocks = append(blocks, blockNewMax)
	ms = append(ms, Mask{iByte: blockNewMax.iByte, mask: blockNewMax.mask, diffSum: float32(blockNewMax.diffSum), ratio: blockNewMax.ratio})
	goto start
}

/*
先计算出所有 BlockNew.list 然后存为 bitmap，之后再利用递归对边区分度
*/
func getPivotMask(bsList []string) (ms []Mask) {
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
		nextBlockNO int     // 下一级的；延续出来的
		nMask       float32 // 留在当前区块数量
		nNextMask   float32 // 分到新分裂区块的数量
	}
	type BlockNew struct {
		mask  byte
		iByte int
		N     int
		list  []Block

		// 评价体系
		newBlock int
		diffSum  float32 // 所有区块 距离中线的分数的和
		ratio    float32
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
	blockNewMax := BlockNew{} // BlockNew{diffSum: math.MaxInt}
outfor:
	for iByte := 0; iByte < lMax; iByte++ {
		for _, mask := range allMask {
			block := BlockNew{
				list: make([]Block, len(bsList)), // escapes to heap
			}
			if len(blocks) > 0 {
				copy(block.list, blocks[len(blocks)-1].list) // 继承最新一次区分好的列表
			}

			mBlockHit := make(map[int]*Hit) // 区块已命中 block
			newBlock := 0                   // 此次循环新增的区块

			for i, bs := range bsList {
				hit := false
				if len(bs) > iByte {
					hit = (mask & bs[iByte]) == mask // 是否命中掩码 mask
				}
				blockNO := block.list[i].NO // 当前 bs 的区块号码
				blockHit := mBlockHit[blockNO]
				if blockHit == nil {
					mBlockHit[blockNO] = &Hit{ // escapes to heap
						hit:     hit,
						blockNO: blockNO,
						nMask:   1,
					}
					continue
				}
				if blockHit.hit == hit {
					blockHit.nMask++ // 留在当前区块数量
					continue
				}
				if blockHit.nextBlockNO == 0 {
					newBlock++
					blockHit.nextBlockNO = fNextBlock()
				}
				blockHit.nNextMask++                    // 分到新分裂区块的数量
				block.list[i].NO = blockHit.nextBlockNO //
				block.list[i].iByte = iByte
			}
			if newBlock == 0 {
				continue
			}

			var diffSum float32 = 0
			for _, hit := range mBlockHit {
				diff := float32(hit.nMask)
				if hit.nMask > hit.nNextMask {
					diff = float32(hit.nNextMask)
				}
				diffSum += diff
			}
			diffSum = diffSum * float32(newBlock)
			// diffSum = float32(newBlock)

			m := map[int]bool{}
			for _, b := range block.list {
				if !m[b.NO] {
					block.N++
					m[b.NO] = true
				}
			}

			if blockNewMax.diffSum < diffSum {
				blockNewMax = BlockNew{
					newBlock: newBlock,
					iByte:    iByte,
					mask:     mask,
					list:     block.list,
					diffSum:  diffSum,
					ratio:    float32(newBlock),
				}
			}

			// 当前区块的数量和 bsList 一样，则表示达到结束条件
			if block.N == len(bsList) {
				ms = append(ms, Mask{iByte: iByte, mask: mask, diffSum: float32(blockNewMax.diffSum), ratio: blockNewMax.ratio})
				blocks = append(blocks, block) //
				// log.Printf("len:%d, blocks:%+v", len(blocks), blocks)
				return
			}
			if diffSum == math.MaxFloat32 {
				// blocks = append(blocks, blockNewMax)
				// ms = append(ms, Mask{iByte: blockNewMax.iByte, mask: blockNewMax.mask})
				// goto start
				break outfor
			}
		}
	}
	blocks = append(blocks, blockNewMax)
	ms = append(ms, Mask{iByte: blockNewMax.iByte, mask: blockNewMax.mask, diffSum: float32(blockNewMax.diffSum), ratio: blockNewMax.ratio})
	goto start
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
		for int(j) < len(idxN) && int(j) != len(idxN) && idxN[j].iByte < i {
			j++
		}
		idxNTable[i] = j
	}
	return
}

func toHash(idxN []iN) (fs []func(k string) int) {
	fs = append(fs, func(k string) int { return 0 })
	for i, x := range idxN {
		f := func(key string) (idx int) {
			idx = fs[i](key)
			if (x.mask & key[x.iByte]) == x.mask {
				idx |= x.iBit
			}
			return
		}
		fs = append(fs, f)
	}

	return
}

//-go:noinline
func hash(key string, idxN []iN) (idx int) {
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

//-go:noinline
func hash2(key string, idxNTable []int16, idxN []iN) (idx int) {
	n := idxNTable[len(key)]
	for _, x := range idxN[:n] {
		if (x.mask & key[x.iByte]) == x.mask {
			idx |= x.iBit
		}
	}
	return
}
func hash2x(key string, idxN []iN) (idx int) {
	for _, x := range idxN {
		if (x.mask & key[x.iByte]) == x.mask {
			idx |= x.iBit
		}
	}
	return
}

//-go:noinline
func hash21(key string, idxNTable []int16, idxN []iN) (idx int) {
	n := idxNTable[len(key)]
	/*
		if (x == a)
		  x= b
		else
		  x= a;
		等价于 x= a ^ b ^ x;
	*/
	for _, x := range idxN[:n] {
		b := int(x.mask & key[x.iByte])
		// if b == 0 { idx |= x.iBit }
		// b = 0 ^ x.iBit ^ b
		b = x.iBit ^ b
		idx |= b
	}
	return
}

//-go:noinline
func byte2bit(b, mask byte) (one byte) {
	// return b & mask
	zero := (mask & b) ^ mask
	one = (1 << zero) & 1
	return
}

// 参考 /usr/local/go/src/runtime/hash64.go: memhashFallback 函数

//-go:noinline
func hash4(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := idxNTable[len(key)]
	for _, x := range idxN[:n] {
		// zero := (x.mask & key[x.iByte]) ^ x.mask
		// one := (1 << zero) & 1
		// idx |= (one << x.i)
		one := int(byte2bit(key[x.iByte], x.mask))
		idx |= one << x.i
	}
	return
}

func hash4x(key string, idxN []iN2) (idx int) {
	for _, x := range idxN {
		mask := x.mask
		b := key[x.iByte]
		one := int(b & mask)
		idx |= one << x.i
	}
	return
}

//-go:noinline
func hash41(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := idxNTable[len(key)]
	for _, x := range idxN[:n] {
		mask := x.mask
		b := key[x.iByte]
		one := int(b & mask)
		idx |= one << x.i
	}
	return
}

// range key 会不会更快一点？
//hash 表就是直接 range key 的

//-go:noinline
func hash3(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := idxNTable[len(key)]
	for _, x := range idxN[:n] {
		idx = int(key[x.iByte])
	}
	return
}

//-go:noinline
func hash31(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := idxNTable[len(key)]
	for i := int16(0); i < n; i++ {
		idx = int(key[i])
	}
	return
}

//-go:noinline
func hash32(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := len(key)
	bs := stringBytes(key)
	var u uint64
	for i := 0; i < n; i += 8 {
		mask := *(*uint64)(unsafe.Pointer(&idxN[i]))
		// bit := *(*int)(pointerOffset(unsafe.Pointer(&idxN), uintptr(i)))
		x := *(*uint64)(unsafe.Pointer(&bs[i])) & mask
		// u |= (x & 0b1) | (0b10 & x >> 4) | (0b1000 & x >> 8)
		u |= (x & 0b1) + 10086
	}
	idx = int(u & 0xffff)
	/*
			0b0000 1000 0001 0010
		+	0b0111 1
		+	0b0011 111111111
		+	0b0001 111111111 1110
			---------------------
			0b1111
	*/

	// idx |= idx >> 32
	// idx |= idx >> 16
	// idx |= idx >> 8
	// idx = idx & 0xff

	return
}

//-go:noinline
func hash33(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := len(key)
	bs := stringBytes(key)
	var u uint64
	for i := 0; i < n; i += 8 {
		mask := *(*uint64)(unsafe.Pointer(&idxN[i]))
		// bit := *(*int)(pointerOffset(unsafe.Pointer(&idxN), uintptr(i)))
		x := *(*uint64)(unsafe.Pointer(&bs[i])) & mask
		u |= (x & 0b1) | (0b10 & x >> 4) | (0b1000 & x >> 8) | (0b1000 & x >> 8)
	}
	idx = int(u & 0xffff)
	return
}

//go:noinline
func hash34(key string, idxNTable []int16, idxN []iN2) (idx int) {
	n := len(key)
	bs := stringBytes(key)
	var u uint64
	for i := 0; i < n; i += 8 {
		mask := *(*uint64)(unsafe.Pointer(&idxN[i]))
		// bit := *(*int)(pointerOffset(unsafe.Pointer(&idxN), uintptr(i)))
		x := *(*uint64)(unsafe.Pointer(&bs[i])) & mask
		u |= (x & 0b1) | (0b10 & x >> 4) | (0b1000 & x >> 8) | (0b1000 & x >> 8)
	}
	idx = int(u & 0xffff)
	return
}

//-go:noinline
func strhasher(p unsafe.Pointer, h uintptr) uintptr {

	return strhash(p, h)
}

//-go:noinline
func hash51(key string, idxNTable []int16, idxN []iN2) (idx int) {
	// return idxNTableFunc[len(key)-1](key, idxN)
	return hashx3(key, idxN)
}

//-go:noinline
func hash52(key string, idxNTable []int16, idxN []iN2) (idx int) {
	// return idxNTableFunc[len(key)-1](key, idxN)
	// return hashx1(key, idxN)

	switch idxNTable[len(key)-1] {
	case 1:
		idx = hashx1(key, idxN)
	case 2:
		idx = hashx2(key, idxN)
	case 3:
		idx = hashx3(key, idxN)
	case 4:
		idx = hashx4(key, idxN)
	case 5:
		idx = hashx5(key, idxN)
	case 6:
		idx = hashx6(key, idxN)
	case 7:
		idx = hashx7(key, idxN)
	case 8:
		idx = hashx8(key, idxN)
	default:
		panic(fmt.Sprintf("getHashParamFunc err:%d", len(key)-1))
	}
	return
}

func hashx8(key string, idxN []iN2) (idx int) {
	x := idxN[0]
	one := int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[1]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[2]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[3]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[4]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[5]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[6]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	x = idxN[7]
	one = int(byte2bit(key[x.iByte], x.mask))
	idx |= (one << x.i)
	return
}
func hashx7(key string, idxN []iN2) (idx int) {
	{
		x := idxN[0]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[1]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[2]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[3]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[4]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[5]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[6]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	return
}
func hashx6(key string, idxN []iN2) (idx int) {
	{
		x := idxN[0]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[1]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[2]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[3]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[4]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[5]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	return
}
func hashx5(key string, idxN []iN2) (idx int) {
	{
		x := idxN[0]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[1]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[2]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[3]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[4]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	return
}
func hashx4(key string, idxN []iN2) (idx int) {
	{
		x := idxN[0]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[1]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[2]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[3]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	return
}
func hashx3(key string, idxN []iN2) (idx int) {
	{
		x := idxN[0]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[1]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[2]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	return
}
func hashx2(key string, idxN []iN2) (idx int) {
	{
		x := idxN[0]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	{
		x := idxN[1]
		one := x.mask & key[x.iByte]
		idx |= int(one) << x.i
	}
	return
}
func hashx1(key string, idxN []iN2) (idx int) {
	x := idxN[0]
	one := x.mask & key[x.iByte]
	idx |= int(one) << x.i
	return
}
