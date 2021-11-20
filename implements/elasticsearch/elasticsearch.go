package elasticsearch

import (
	"bytes"
	"errors"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	jsoniter "github.com/json-iterator/go"
	"github.com/mutou1225/go-frame/logger"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EsMap map[string]interface{}

var (
	esClient = make(map[string]*elasticsearch.Client)
	esmutex  sync.Mutex
)

func InitEsClient(host string) error {
	if host == "" {
		return errors.New("InitEsClient() Err: host empty")
	}

	esmutex.Lock()
	defer esmutex.Unlock()

	if client, ok := esClient[host]; ok && client != nil {
		return nil
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://" + host,
		},
		/*
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: 10 * time.Millisecond,
			},
		*/
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second, // 连接超时时间
				KeepAlive: 30 * time.Second, // 连接保持超时时间
			}).DialContext,
			MaxIdleConns:        200,              // 最大连接数,默认0无穷大
			MaxIdleConnsPerHost: 200,              // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
			IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接
		},
		Logger: EsLogger{},
	}

	esConn, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("elasticsearch.NewClient() response: %s", err)
		return err
	}
	esClient[host] = esConn

	res, err := esConn.Info()
	if err != nil {
		log.Printf("elasticsearch.NewClient() response: %s", err)
		return err
	}
	defer res.Body.Close()

	// Check response status
	if res.IsError() {
		log.Printf("elasticsearch.NewClient() Error: %s", res.String())
		return err
	}

	// Deserialize the response into a map.
	r := EsMap{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
		return err
	}
	// Print client and server version numbers.
	log.Printf("ES Client: %s", elasticsearch.Version)
	log.Printf("ES Server: %s", r["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))

	return nil
}

func CloseEsClient(host string) {
	esmutex.Lock()
	defer esmutex.Unlock()

	if _, ok := esClient[host]; ok {
		esClient[host] = nil
	}
}

// 获取es连接，操作完后无需close
func GetEsClient(host string) (*elasticsearch.Client, error) {
	if host == "" {
		return nil, errors.New("host empty")
	} else if esClient == nil {
		if err := InitEsClient(host); err != nil {
			return nil, err
		} else if esClient == nil {
			return nil, errors.New("elasticsearch uninitialized")
		}
	}

	esmutex.Lock()
	defer esmutex.Unlock()

	if client, ok := esClient[host]; ok && client != nil {
		return client, nil
	} else {
		if err := InitEsClient(host); err != nil {
			return nil, err
		} else if client, ok := esClient[host]; ok && client != nil {
			return client, nil
		}
	}

	return nil, errors.New("client uninitialized")
}

func RespIsError(res *esapi.Response) bool {
	if res.IsError() {
		var e map[string]interface{}
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			logger.PrintError("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			logger.PrintError("Es Resp: [%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
		return true
	}

	return false
}

// 返回的[]interface{}结构为 map[string]interface{}
func GetRespSource(resp *esapi.Response) ([]interface{}, error) {
	var r map[string]interface{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		logger.PrintError("Error parsing the response body: %s", err.Error())
		return nil, err
	}

	// Print the response status, number of results, and request duration.
	hist := r["hits"].(map[string]interface{})["hits"].([]interface{})
	logger.PrintInfo(
		"Es Resp: [%s] total[%d] hits[%d]; took: %dms",
		resp.Status(),
		int(r["hits"].(map[string]interface{})["total"].(float64)),
		len(hist),
		int(r["took"].(float64)),
	)

	ret := make([]interface{}, 0, len(hist))
	for _, data := range hist {
		d := data.(map[string]interface{})["_source"]
		ret = append(ret, d)
	}

	return ret, nil
}

type EsLogger struct{}

func (l EsLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	logger.PrintInfo("ES: %s %s [status:%d request:%d]",
		req.Method,
		req.URL.String(),
		resStatusCode(res),
		dur.Truncate(time.Millisecond),
	)

	if l.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			buf.ReadFrom(b)
		} else {
			buf.ReadFrom(req.Body)
		}
		logger.PrintInfo("ES Request: %s", buf.String())
	}

	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer res.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(res.Body)
		logger.PrintInfo("ES Response: %s", buf.String())
	}

	if err != nil {
		logger.PrintError("ES ERROR: %v", err)
	}

	return nil
}

// RequestBodyEnabled makes the client pass a copy of request body to the logger.
func (l EsLogger) RequestBodyEnabled() bool {
	return true
}

// ResponseBodyEnabled makes the client pass a copy of response body to the logger.
func (l EsLogger) ResponseBodyEnabled() bool {
	return false
}

func resStatusCode(res *http.Response) int {
	if res == nil {
		return -1
	}
	return res.StatusCode
}
