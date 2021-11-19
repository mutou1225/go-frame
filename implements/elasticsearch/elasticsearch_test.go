package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"
)

func TestEsClient(t *testing.T) {
	esHost := "EVA_ES_HOST:9200"
	if err := InitEsClient(esHost); err != nil {
		t.Errorf("InitEsClient() error:%s", err.Error())
	}

	esClient, err := GetEsClient(esHost)
	if err != nil {
		t.Errorf("GetEsClient() Err: %s", err.Error())
	}

	query := EsMap{
		"track_total_hits": true,
		"from":             0,
		"size":             10,
		"query": EsMap{
			"bool": EsMap{
				"must": []EsMap{
					{"match": EsMap{"Feva_status": 1}}, // 估价状态
					{"match": EsMap{"Fstatus": 1}},     // 产品状态
					{"match": EsMap{"Fvalid": 1}},      // 模板状态
				},
			},
		},
		"sort": []EsMap{{"Fproduct_id": "desc"}},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		t.Errorf("json.NewEncoder() Err: %s", err.Error())
	}

	// Perform the search request.
	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("eva_template_product"),
		esClient.Search.WithBody(&buf),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
		esClient.Search.WithDocumentType("_doc"),
	)
	if err != nil {
		t.Errorf("esClient.Search() Err: %s", err.Error())
	} else if res == nil {
		t.Errorf("esClient.Search() res <nil>")
	}
	defer res.Body.Close()
	log.Printf("res: %s", res.Status())

	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll() Err: %s", err.Error())
		}

		esResult := EsMap{}
		if err = json.Unmarshal(body, &esResult); err != nil {
			t.Errorf("json.Unmarshal() Err: %s", err.Error())
		}
		log.Printf("esResult: %+v", esResult)
	} else {
		t.Errorf("res.StatusCode[%d] not 200", res.StatusCode)
	}
}

func Benchmark_EsClient(b *testing.B) {
	esHost := "EVA_ES_HOST:9200"
	if err := InitEsClient(esHost); err != nil {
		b.Errorf("InitEsClient() error:%s", err.Error())
	}

	for i := 0; i < b.N; i++ {
		esClient, err := GetEsClient(esHost)
		if err != nil {
			b.Errorf("GetEsClient() Err: %s", err.Error())
		}

		query := EsMap{
			"track_total_hits": true,
			"from":             0,
			"size":             10,
			"query": EsMap{
				"bool": EsMap{
					"must": []EsMap{
						{"match": EsMap{"Feva_status": 1}}, // 估价状态
						{"match": EsMap{"Fstatus": 1}},     // 产品状态
						{"match": EsMap{"Fvalid": 1}},      // 模板状态
					},
				},
			},
			"sort": []EsMap{{"Fproduct_id": "desc"}},
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			b.Errorf("json.NewEncoder() Err: %s", err.Error())
		}

		// Perform the search request.
		res, err := esClient.Search(
			esClient.Search.WithContext(context.Background()),
			esClient.Search.WithIndex("eva_template_product"),
			esClient.Search.WithBody(&buf),
			esClient.Search.WithTrackTotalHits(true),
			esClient.Search.WithPretty(),
			esClient.Search.WithDocumentType("_doc"),
		)
		if err != nil {
			b.Errorf("esClient.Search() Err: %s", err.Error())
		} else if res == nil {
			b.Errorf("esClient.Search() res <nil>")
		}
		res.Body.Close()
		log.Printf("res: %s", res.Status())
	}

	CloseEsClient(esHost)
}
