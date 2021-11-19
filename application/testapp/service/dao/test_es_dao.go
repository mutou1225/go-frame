package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"eva_services_go/config"
	es "eva_services_go/implements/elasticsearch"
	"eva_services_go/logger"
	"fmt"
	"io/ioutil"
)

type EsResq struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			StrId  string `json:"_id"`
			Source struct {
				ClassId     int    `json:"Fclass_id"`
				ClassName   string `json:"Fclass_name"`
				ProductId   int    `json:"Fproduct_id"`
				ProductName string `json:"Fproduct_name"`
				BrandId     int    `json:"Fbrand_id"`
				BrandName   string `json:"Fbrand_name"`
				PicId       string `json:"Fpic_id"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func GetInfoFromES(pageIndex, pageSize int) (*EsResq, error) {
	esClient, err := es.GetEsClient(config.GetESHost())
	if err != nil {
		logger.PrintInfo("GetEsClient() Err: %s", err.Error())
	}

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]map[string]string{
					{"match": {"Fis_upper": "1"}},
					{"match": {"Fplatform_type": "1"}},
					{"match": {"Fvalid": "1"}},
					{"match": {"Fshow_flag": "1"}},
				},
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		logger.PrintError("json.NewEncoder() Err: %s", err.Error())
	}
	logger.PrintInfo("buf: %s", buf)

	// Perform the search request.
	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("eva_platform_product"),
		esClient.Search.WithFrom(pageIndex),
		esClient.Search.WithSize(pageSize),
		esClient.Search.WithSource("Fproduct_id", "Fproduct_name", "Fbrand_id", "Fbrand_name", "Fclass_id", "Fclass_name", "Fpic_id"),
		esClient.Search.WithBody(&buf),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		logger.PrintError("es.Search() Err: %s", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.PrintInfo("Err: %s", err.Error())
			return nil, err
		}

		esResult := EsResq{}
		if err = json.Unmarshal(body, &esResult); err != nil {
			logger.PrintInfo("Err: %s", err.Error())
			return nil, err
		}
		logger.PrintInfo("esResult: %+v", esResult)

		return &esResult, nil
	}

	return nil, errors.New(fmt.Sprintf("StatusCode: %d", res.StatusCode))
}
