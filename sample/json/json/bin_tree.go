package json

import (
	"log"
	"sort"
)

func getHashFuncU64(idxs []int) (f func(key []byte) (idx int)) {
	// 可以提前聚合成 byte ，然后打表合并 bit 增速最高 8 倍？

	type N struct {
		iByte int // []byte的偏移量
		mask  byte
		iBit  int
	}
	idxN := make([]N, len(idxs))
	for i, idx := range idxs {
		iKey := idx / 8 // key 的偏移量
		iBit := idx % 8 // byte 内部偏移量
		idxN[i] = N{
			iByte: iKey, // key 的偏移量
			mask:  1 << iBit,
			iBit:  1 << i,
		}
	}
	f = func(key []byte) (idx int) {
		for i, x := range idxN {
			if i >= len(key) {
				break
			}
			if x.mask&key[x.iByte] > 0 {
				idx |= x.iBit
			}
		}
		return
	}
	return
}

type iN struct {
	iByte int // []byte的偏移量
	mask  byte
	iBit  int
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

func logicalHash(bsList [][]byte) (idxN []iN, idxRet []int) {
	var idxsRet [][]int
	if len(bsList) <= 1 {
		idxN = []iN{{
			iByte: 0, // key 的偏移量
			mask:  1,
			iBit:  1,
		}}
		return
	}

	idxsRet = divide(bsList)
	log.Printf("---idxsRet---:%+v", idxsRet)
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

// 只能通过贪心算法加快计算速度，拿个不是那么好的结果
// 用 4bit 或 8 bit 来搞，一次能分的类型最多的获胜
// 从 1bit 追到 8bit(64bit)，只为找能能对半分的
func divide(bsList [][]byte) (idxsRet [][]int) {
	// PrintKeys(bsList)
	if len(bsList) == 1 {
		return [][]int{{}}
	}
	lMax := len(bsList[0])
	for _, bs := range bsList {
		if lMax < len(bs) {
			lMax = len(bs)
		}
	}
	n1Bits := make([]uint32, 8*lMax) // bit位为 1 的个数
	for _, bs := range bsList {
		for i, b := range bs {
			idx := i * 8
			if idx >= len(n1Bits) {
				break
			}
			n1Bit8 := n1Bits[idx:]
			for j := 0; j < 8; j++ {
				n1Bit8[j] += (uint32(b) >> (7 - j)) & 1 // 小端
			}
		}
	}

	//选 8 个最接近中间值的
	type Mid struct {
		i     int
		diff  float32
		n1Bit uint32 // bit 位为 1 的数量
	}
	l := uint32(len(bsList))
	mid := float32(l) / 2 //获取中值
	N := 1                // 贪心算法的余量，一次可以贪 N 个，最后选最优
	midList := make([]Mid, 0, 8)
	for i, n1Bit := range n1Bits {
		if n1Bit == 0 || n1Bit == l {
			continue
		}

		diff := mid - float32(n1Bit)
		if diff < 0 {
			diff = -diff
		}
		m := Mid{
			i:     i,
			diff:  diff,
			n1Bit: n1Bit,
		}
		if len(midList) < cap(midList) {
			midList = append(midList, m)
			if len(midList) == cap(midList) {
				sort.SliceStable(midList, func(i, j int) bool { return midList[i].diff < midList[j].diff })
			}
			continue
		}

		for i := range midList {
			if midList[i].diff > m.diff {
				midList = append(midList[:i+1], midList[i:len(midList)-1]...)
				midList[i] = m
				break
			}
		}
	}
	if len(bsList) == 2 {
		for _, m := range midList {
			idxsRet = append(idxsRet, []int{m.i})
		}
		return
	}

	for _, m := range midList {
		// 先按比特位分成两拨
		bsLeft := make([][]byte, 0, m.n1Bit)
		bsRight := make([][]byte, 0, l-m.n1Bit)
		for _, bs := range bsList {
			idxX := m.i / 8
			idxY := m.i % 8
			if len(bs) <= idxX {
				bsLeft = append(bsLeft, bs)
				continue
			}
			b := bs[idxX]
			if (uint32(b)>>(7-idxY))&1 == 0 {
				bsLeft = append(bsLeft, bs)
			} else {
				bsRight = append(bsRight, bs)
			}
		}
		idxListL := divide(bsLeft)
		idxListR := divide(bsRight)
		// idxsRet1 := [][]int{}
		for _, idxL := range idxListL {
			for _, idxR := range idxListR {
				idxMin, idxMax := idxL, idxR
				if len(idxL) > len(idxR) {
					idxMin, idxMax = idxR, idxL
				}
				idx := idxMin[:len(idxMin):len(idxMin)]
				for _, i := range idxMax {
					bFound := false
					for _, j := range idxMin {
						if j == i {
							bFound = true
							break
						}
					}
					if !bFound {
						idx = append(idx, i)
					}
				}
				idx = append(idx, int(m.i))
				idxsRet = append(idxsRet, idx)
				// idxsRet1 = append(idxsRet1, idx)
			}
		}
		// log.Printf("idxsRet1:%+v", idxsRet1)
		// idxsRet = append(idxsRet, idxsRet1...)
	}
	// if len(idxsRet) == 0 {
	// 	// log.Printf("idxsRet:%+v", idxsRet)
	// 	panic("000")
	// }

	idxs := make([][]int, 0, N) // 按顺序，第一个最小
	for _, idx := range idxsRet {
		if len(idxs) < cap(idxs) {
			idxs = append(idxs, idx)
			if len(idxs) == cap(idxs) {
				sort.SliceStable(idxs, func(i, j int) bool { return len(idxs[i]) < len(idxs[j]) })
			}
			continue
		}
		for i := range idxs {
			if len(idxs[i]) > len(idx) {
				idxs = append(idxs[:i+1], idxs[i:len(idxs)-1]...)
				idxs[i] = idx
				break
			}
		}
	}
	idxsRet = idxs
	return
}

//根据状态数 返回 bit 数
func getNBit(nStatus int) (nBit int) {
	x := 2
	nBit = 1
	for x < nStatus {
		x *= 2
		nBit++
	}
	return
}
func getNLen(nBit int) (nStatus int) {
	nStatus = 1
	for i := 0; i < nBit; i++ {
		nStatus *= 2
	}
	return
}
