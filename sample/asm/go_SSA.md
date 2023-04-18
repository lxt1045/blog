
为提高性能-静态替换！！！

https://draveness.me/golang/docs/part1-prerequisite/ch02-compile/golang-ir-ssa/
https://oftime.net/2021/02/14/ssa/
https://tonybai.com/2022/10/21/understand-go-ssa-by-example/



https://zh.wikipedia.org/zh-cn/%E9%9D%99%E6%80%81%E5%8D%95%E8%B5%8B%E5%80%BC%E5%BD%A2%E5%BC%8F
https://www.cnblogs.com/crossain/p/13711939.html
https://golang.design/under-the-hood/zh-cn/part3tools/ch11compile/ssa/
https://gocompiler.shizhz.me/10.-golang-bian-yi-qi-han-shu-bian-yi-ji-dao-chu/10.2.1-ssa
https://bbs.huaweicloud.com/blogs/227535




/usr/local/go/src/cmd/compile/internal/ssagen/ssa.go
```go
// buildssa builds an SSA function for fn.
// worker indicates which of the backend workers is doing the processing.
func buildssa(fn *ir.Func, worker int) *ssa.Func {
}
```


/usr/local/go/src/cmd/compile/internal/ssa/rewriteAMD64.go
```go
func rewriteValueAMD64(v *Value) bool {
}
```