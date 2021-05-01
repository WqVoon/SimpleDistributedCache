package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 默认链接前缀，用于区分主机上的服务是否为当前的缓存服务
const defaultBasePath = "/CacheServe/"

// HTTP 服务端结构体，实现了 ServeHTTP 方法用于根据接口提供缓存值
type HTTPPool struct {
	// 用于标识服务
	self string
	// 用于根据接口区分服务，其值等于 defaultBasePath
	basePath string
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
