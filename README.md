## MINI-DistributedCache

-   缓存替换策略实现了并发安全的 LRU 和 LRU-k 算法(默认替换策略是 LRU-k)
-   service 是对 Cache 做了一个上层的封装
    -   单机可以实例化多个服务，多次调用 `NewService` 输入不同 name 即可
    -   内部使用 groups 记录 name 对应的 Service 
-   server 通过实现 `http.ServeServeHTTP` 进行挂载
-   Map 实现节点添加、删除、请求的转发
    -   每个真实节点复制生成了多个虚拟节点
    -   利用一致性哈希，将空间划分为 0~2^32 - 1 的哈希环
    -   将虚拟节点映射到哈希环上
    -   收到请求后，计算请求 key 的哈希值，顺时针寻找距其最近的节点