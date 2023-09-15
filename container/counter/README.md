# LinkCounter
### 一个限制长度，自动回撤的优先队列 + BitMap实现的关联列表的容器

用途：假设有8万个ID，以及无法估量的Key，每次请求服务，会有一个Key携带一个ID以及一个Hit值。服务将累积每个Key的Hit值，并实时统计前1万Hit值最高的Key，以及此Key携带的所有ID。由于内存有限，服务中实际最多保存前2万Hit值的Key。则有如下示例：

```
lc := NewLinkCounter(10000, 20000)
lc.Add("a", 2, "x")
lc.Add("a", 1, "y")
lc.Add("b", 1, "y")
// ...
lc.Range(func(key string, hits int64, linkIds []string) bool {
    fmt.Printf("key: %s, hits: %d, linkIds: %v\n", key, hits, linkIds)
    return true
})
```