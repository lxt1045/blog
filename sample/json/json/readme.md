当前存在的问题：
1. pointer、slice 的 cache 的 tag 的名字相同的的时候，会有冲突
2. slice 套 slice，pointer slice