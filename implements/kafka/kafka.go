package kafka

import (
	"errors"
	"eva_services_go/logger"
	"fmt"
	"github.com/Shopify/sarama"
	"sync"
)

var (
	kfkProducerClient = make(map[string]sarama.Client)
	kfkConsumerClient = make(map[string]sarama.Consumer)
)

func InitKafkaProducer(host string) error {
	if host == "" {
		return errors.New("host empty")
	}

	if client, ok := kfkProducerClient[host]; ok && client != nil {
		return nil
	}

	config := sarama.NewConfig() //实例化个sarama的Config
	//是否开启消息发送成功后通知 successes channel, 如果打开了Return.Successes配置，而又没有producer.Successes()提取，那么Successes()这个chan消息会被写满。
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner //随机分区器
	//config.Producer.RequiredAcks = sarama.WaitForAll
	client, err := sarama.NewClient([]string{host}, config) //初始化客户端
	if err != nil {
		logger.PrintError("sarama.NewClient() Err: %s", err.Error())
		kfkProducerClient[host] = nil
		return err
	}
	kfkProducerClient[host] = client
	return nil
}

func CloseKafkaProducer(host ...string) {
	if len(host) == 0 {
		for h, c := range kfkProducerClient {
			if c != nil {
				c.Close()
				kfkProducerClient[h] = nil
			}
		}
	} else {
		for _, h := range host {
			if c, ok := kfkProducerClient[h]; ok && c != nil {
				c.Close()
				kfkProducerClient[h] = nil
			}
		}
	}
}

func GetProducerClient(host string) (sarama.Client, error) {
	if kfkProducerClient == nil {
		return nil, errors.New("kafka uninitialized")
	}

	if host == "" {
		return nil, errors.New("host empty")
	}

	if client, ok := kfkProducerClient[host]; ok && client != nil {
		return client, nil
	}

	return nil, errors.New("client uninitialized")
}

// 同步模式生产者
func SyncProducerMessage(topic, key, msg *string, host string) error {
	client, err := GetProducerClient(host)
	if err != nil {
		return err
	}

	prdMsg, err := newProducerMessage(topic, key, msg)
	if err != nil {
		logger.PrintError("newProducerMessage() Err: %s", err.Error())
		return err
	}

	producerSync, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		logger.PrintError("NewSyncProducerFromClient() Err: %s", err.Error())
		return err
	}
	defer producerSync.Close()

	partition, offset, err := producerSync.SendMessage(prdMsg)
	if err != nil {
		logger.PrintError("unable to produce message: %s", err.Error())
	}
	logger.PrintInfo("partition", partition)
	logger.PrintInfo("offset", offset)

	return nil
}

// 异步模式生产者
func AsyncProducerMessage(topic, key, msg *string, host string) error {
	client, err := GetProducerClient(host)
	if err != nil {
		return err
	}

	prdMsg, err := newProducerMessage(topic, key, msg)
	if err != nil {
		logger.PrintError("newProducerMessage() Err: %s", err.Error())
		return err
	}

	producerAsync, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		logger.PrintError("NewAsyncProducerFromClient() Err: %s", err.Error())
		return err
	}

	producerAsync.Input() <- prdMsg

	go func() {
		defer producerAsync.Close()

		// wait response
		select {
		case msg := <-producerAsync.Successes():
			logger.PrintInfo("Produced message successes: [%s]", msg.Value)
			break
		case err := <-producerAsync.Errors():
			logger.PrintError("Produced message failure: %s", err.Error())
			break
		default:
			logger.PrintInfo("Produced message default")
			break
		}
	}()

	return nil
}

func newProducerMessage(topic, key, msg *string) (*sarama.ProducerMessage, error) {
	if topic == nil || *topic == "" {
		return nil, errors.New("topic is nil or empty!")
	}

	if msg == nil {
		return nil, errors.New("msg is nil!")
	}

	prdMsg := sarama.ProducerMessage{}
	prdMsg.Topic = *topic
	prdMsg.Value = sarama.StringEncoder(*msg)
	if key != nil && *key != "" {
		prdMsg.Key = sarama.StringEncoder(*key)
	}

	return &prdMsg, nil
}

func InitKafkaConsumer(host string) error {
	if host == "" {
		return errors.New("host empty")
	}

	if client, ok := kfkConsumerClient[host]; ok && client != nil {
		return nil
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = false
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	//config.Version = sarama.V0_10_2_0

	client, err := sarama.NewConsumer([]string{host}, config)
	if err != nil {
		logger.PrintError("sarama.NewClient() Err: %s", err.Error())
		kfkConsumerClient[host] = nil
		return err
	}
	kfkConsumerClient[host] = client
	return nil
}

func CloseKafkaConsumer(host ...string) {
	if len(host) == 0 {
		for h, c := range kfkConsumerClient {
			if c != nil {
				c.Close()
				kfkConsumerClient[h] = nil
			}
		}
	} else {
		for _, h := range host {
			if c, ok := kfkConsumerClient[h]; ok && c != nil {
				c.Close()
				kfkConsumerClient[h] = nil
			}
		}
	}
}

func GetConsumeClient(host string) (sarama.Consumer, error) {
	if kfkConsumerClient == nil {
		return nil, errors.New("kafka uninitialized")
	}

	if host == "" {
		return nil, errors.New("host empty")
	}

	if client, ok := kfkConsumerClient[host]; ok && client != nil {
		return client, nil
	}

	return nil, errors.New("client uninitialized")
}

type ConsumerHandler struct {
	Name string
}

// Offset can be a literal offset, or OffsetNewest or OffsetOldest
// 使用: 在业务逻辑中通过 consumer.Messages() 获取信息，如下例子
/*
	for i := 0; i < 10; i++ {
		select {
		case message := <- consumer.Messages():
			...
		case err := <-consumer.Errors():
			...
		}
	}
*/
func GetConsumePartition(host string, topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	client, err := GetConsumeClient(host)
	if err != nil {
		return nil, err
	}

	consumer, err := client.ConsumePartition(topic, partition, offset)
	if err != nil {
		logger.PrintError("ConsumePartition() Err: %s", err.Error())
		return nil, err
	}

	return consumer, nil
}

func ConsumeGroupMessage(host string, groupID string, handler ConsumerHandler, topics ...string) error {
	/*
		client, err := GetConsumeClient(host)
		if err != nil {
			return err
		}

		group, err := sarama.NewConsumerGroupFromClient(groupID, client)
		if err != nil {
			return err
		}
		defer group.Close()

		for {
			err := group.Consume(context.Background(), topics, handler)
			if err != nil {
				logger.PrintError("group.Consume() Err: %s", err.Error())
			}
		}
	*/

	return nil
}

func (ConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h ConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("%s Message topic:%q partition:%d offset:%d  value:%s\n", h.Name, msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		// 手动确认消息
		sess.MarkMessage(msg, "")
	}
	return nil
}

func handleErrors(group *sarama.ConsumerGroup, wg *sync.WaitGroup) {
	wg.Done()
	for err := range (*group).Errors() {
		fmt.Println("ERROR", err)
	}
}
