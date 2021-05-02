package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// 默认链接前缀，用于区分主机上的服务是否为当前的缓存服务
const (
	defaultBasePath = "/CacheServer/"
	defaultReplicas = 50
)

/*
HTTP 服务端结构体，实现了 ServeHTTP 方法用于根据接口提供缓存值；
同时实现了 PeerPicker 接口
*/
type HTTPPool struct {
	// 用于标识服务，方便日志输出
	self string
	// 用于根据接口区分服务，其值等于 defaultBasePath
	basePath string
	mu       sync.Mutex
	// 实现了一致性哈希的映射，可根据键值找到对应的节点名
	peers *Map
	// 存储节点名到 httpGetter 的映射
	httpGetters map[string]*httpGetter
}

func makeHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (self *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s]%s", self.self, fmt.Sprintf(format, v...))
}

func (self *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.Log("%s %s", r.Method, r.URL.Path)

	// 仅为 /CacheNode/* 的请求提供服务
	if !strings.HasPrefix(r.URL.Path, self.basePath) {
		http.Error(w, "HTTPPool serving unexpacted path:"+r.URL.Path, http.StatusBadRequest)
		return
	}
	// 请求的格式为 /<basePath>/<groupName>/<key>，这里的 parts 为 []string{<groupName>, <key>}
	parts := strings.SplitN(r.URL.Path[len(self.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad reqeust", http.StatusBadRequest)
		return
	}

	groupName, key := parts[0], parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	value, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value.ByteSlice())
}

/*
延迟赋值 self.peers 和 self.httpGetters，并利用传递的 peers 参数来初始化之
*/
func (self *HTTPPool) Set(peers ...string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.peers = makeMap(defaultReplicas, nil)
	self.peers.Add(peers...)
	self.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		self.httpGetters[peer] = &httpGetter{
			baseURL: peer + self.basePath,
		}
	}
}

// 为 HTTPPool 实现 PeerPicker 接口，使其根据 key 返回对应的 httpGetter
func (self *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if peer := self.peers.Get(key); peer != "" && peer != self.self {
		self.Log("Pick peer %s", peer)
		return self.httpGetters[peer], true
	}
	return nil, false
}

// httpGetter 是 PeerGetter 在 HTTP 协议上的一个实现
type httpGetter struct {
	baseURL string
}

/*
利用 self.baseURL，group 和 key 来拼接请求链接 u，
并调用 http.Get(u) 来向某个 Group 获取数据
*/
func (self *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		self.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server returned: %v\n", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error when reading response body: %v", err)
	}

	return bytes, nil
}
