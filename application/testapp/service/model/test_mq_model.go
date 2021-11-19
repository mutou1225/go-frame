package model

import (
	"encoding/json"
	"eva_services_go/application/testapp/appstorage"
	"eva_services_go/implements/rabbitmq"
	"eva_services_go/logger"
	"sync"
)

const (
	testExchangeName = "test-exchange"
	testQueueName    = "test-queeu"
)

type testMQ struct{}

func (r testMQ) OnError(err error) {
	logger.PrintError("MQOperation Err: %s", err.Error())
}

func (r testMQ) QueueName() string {
	return testQueueName
}

func (r testMQ) ConsumeName() string {
	return "test-go"
}

func (r testMQ) RouterKey() string {
	return ""
}

func TestMQFunc(wg *sync.WaitGroup, endChan chan struct{}) {
	defer wg.Done()

	client, err := appstorage.GetRabbitMQClient()
	if err != nil {
		logger.PrintError("rabbitmq.NewRabbitMQ() Err: %s", err.Error())
		return
	}

	channel, err := client.ConnAndChannel()
	if err != nil {
		logger.PrintError("rabbitmq.ConnAndChannel() Err: %s", err.Error())
		return
	}

	if err := client.CreateExchange(channel, testExchangeName, rabbitmq.MQKindFanout); err != nil {
		logger.PrintError("rabbitmq.CreateExchange() Err: %s", err.Error())
		return
	}

	if err := client.CreateQueue(channel, testQueueName); err != nil {
		logger.PrintError("rabbitmq.CreateQueue() Err: %s", err.Error())
		return
	}

	if err := client.QueueBind(channel, testQueueName, "", testExchangeName); err != nil {
		logger.PrintError("rabbitmq.CreateExchange() Err: %s", err.Error())
		return
	}

	// 监听退出信号
	go func() {
		<-endChan
		client.EndConsumeQueue()
		logger.PrintInfo("get end chan")
		return
	}()

	client.ConsumeQueue(channel, testMQ{}, 2)
}

type testMQInfo struct {
	Id   int
	Name string
}

func (r testMQ) OnReceive(msgData []byte, workId int) (retAck bool) {
	retAck = true
	logger.PrintInfo("OnReceive msg: %+v ", string(msgData))

	var msgVal testMQInfo
	if err := json.Unmarshal(msgData, &msgVal); err != nil {
		logger.PrintError("json.Unmarshal() Err: %s", err.Error())
		return
	}
	logger.PrintInfo("revMsg: %+v", msgVal)

	// 业务逻辑操作 。。。

	return
}
