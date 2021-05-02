package main

// 利用一个 key 来根据一致性哈希算法找到对应的 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 用于从对应的组中查找缓存值
type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}
