package http

import (
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	client *http.Client
	once   sync.Once
)

func CreateHTTPClient() *http.Client {
	// 使用单例创建client
	once.Do(func() {
		client = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second, // 连接超时时间
					KeepAlive: 30 * time.Second, // 连接保持超时时间
				}).DialContext,
				MaxIdleConns:        2000,             // 最大连接数,默认0无穷大
				MaxIdleConnsPerHost: 2000,             // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
				IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接
			},
		}
	})
	return client
}
