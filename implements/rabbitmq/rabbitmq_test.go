package rabbitmq

import (
	jsoniter "github.com/json-iterator/go"
	"log"
	"os"
	"testing"
	"time"
)

var (
	client *RabbitMQ
)

const (
	pushMsg = "RabbitMQ Test"
)

func setup() {
	clientMQ, err := NewRabbitMQ("hsb", "hsb.com",
		"EVA_MQ_HOST", 5672, "eva_vhost")
	if err != nil {
		log.Panicf("NewRabbitMQ() Err: %s", err.Error())
	}

	client = clientMQ
}

func teardown() {
	if client != nil {
		client.Close()
	}
}

type testMqType struct{}

func (d testMqType) OnError(err error) {
	log.Printf("MQOperation Err: %s", err.Error())
}

func (d testMqType) QueueName() string {
	return "go-mq-test-queue"
}

func (d testMqType) ConsumeName() string {
	return "go-mq-test-consume"
}

func (d testMqType) RouterKey() string {
	return ""
}

func (d testMqType) ExchangeKey() string {
	return "go-mq-test-exchange"
}

func (d testMqType) OnReceive(msgData []byte, workId int) (retAck bool) {
	retAck = true

	log.Printf("OnReceive: %s", string(msgData))

	var receive string
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(msgData, &receive); err != nil {
		log.Panicf("json.Unmarshal() Err: %s", err.Error())
	}

	if pushMsg != receive {
		log.Panicf("OnReceive() pushMsg != receive")
	}

	return
}

func TestRabbitMQApi(t *testing.T) {
	tMq := testMqType{}
	channel, err := client.ConnAndChannel()
	if err != nil {
		t.Errorf("rabbitmq.ConnAndChannel() Err: %s", err.Error())
	}
	defer channel.Close()

	if err := client.CreateExchange(channel, tMq.ExchangeKey(), MQKindFanout); err != nil {
		t.Errorf("rabbitmq.CreateExchange() Err: %s", err.Error())
		return
	}

	if err := client.CreateQueue(channel, tMq.QueueName()); err != nil {
		t.Errorf("rabbitmq.CreateQueue() Err: %s", err.Error())
		return
	}

	if err := client.QueueBind(channel, tMq.QueueName(), tMq.RouterKey(), tMq.ExchangeKey()); err != nil {
		t.Errorf("rabbitmq.QueueBind() Err: %s", err.Error())
		return
	}

	// 编解码都用 json
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonBytes, err := json.Marshal(pushMsg)
	if err != nil {
		t.Errorf("json.Marshal() Err: %s", err.Error())
		return
	}

	if err := client.PublishQueue(channel, tMq.ExchangeKey(), tMq.RouterKey(), string(jsonBytes)); err != nil {
		t.Errorf("rabbitmq.PublishQueue() Err: %s", err.Error())
		return
	}

	channelConsume, err := client.ConnAndChannel()
	if err != nil {
		t.Errorf("rabbitmq.ConnAndChannel() Err: %s", err.Error())
		return
	}
	defer channelConsume.Close()

	// if block is true, must be called: client.EndConsumeQueue()
	go func() {
		time.Sleep(time.Second * 1)
		client.EndConsumeQueue()
	}()

	err = client.ConsumeQueue(channelConsume, tMq, 1)
	if err != nil {
		t.Errorf("rabbitmq.ConsumeQueue() Err: %s", err.Error())
		return
	}

	err = client.DeleteQueue(channel, tMq.QueueName())
	if err != nil {
		t.Errorf("rabbitmq.DeleteQueue() Err: %s", err.Error())
		return
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
