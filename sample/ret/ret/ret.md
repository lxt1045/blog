
Ret 的方式有个很大的缺点，就是不能鞋底返回值，因为 goleng 编译器默认编译成 APIinternal API 接口形式，需要通过寄存器返回值，但是对应的映射需要在 DWRF 中才有，现在我还没有解析改端的能力


另一个是返回的函数不能内联，通过一下注释
//go:noinline

不过前景倒是挺乐观的！