package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

type MQKind string

const (
	MQKindDirect  MQKind = "direct"
	MQKindFanout  MQKind = "fanout"
	MQKindTopic   MQKind = "topic"
	MQKindHeaders MQKind = "headers"
)

// 提供的操作
type MQOperation interface {
	ConnAndChannel() (*amqp.Channel, error)                                    // 获取可用的channel
	CreateQueue(*amqp.Channel, string) error                                   // 创建一个queue队列
	CreateQueueWithArgs(*amqp.Channel, string, amqp.Table) error               // 创建一个queue队列
	CreateExchange(*amqp.Channel, string, MQKind) error                        // 创建一个Exchange
	CreateExchangeWithArgs(*amqp.Channel, string, MQKind, amqp.Table) error    // 创建一个Exchange
	QueueBind(*amqp.Channel, string, string, string) error                     // 队列绑定
	QueueBindWithArgs(*amqp.Channel, string, string, string, amqp.Table) error // 队列绑定
	DeleteQueue(*amqp.Channel, string) error                                   // 删除一个queue队列
	PublishQueue(*amqp.Channel, string, string, string) error                  // 发布消息到队列
	ConsumeQueue(*amqp.Channel, MQConsume, int) error                          // 取出消息消费
	EndConsumeQueue()                                                          // 退出消息消费
	ReConsume(string, string, string) error                                    // 退回消息
	GetReadyCount(*amqp.Channel, string) (int, error)                          // 统计正在队列中准备且还未消费的数据
	GetConsumCount(*amqp.Channel, string) (int, error)                         // 获取到队列中正在消费的数据，这里指的是正在有多少数据被消费
	QueuePurge(string) (int, error)                                            // 清理队列
}

//消费者定义
type MQConsume interface {
	OnError(error)                          // 处理遇到的错误，当RabbitMQ对象发生了错误，他需要告诉接收者处理错误
	OnReceive(body []byte, workId int) bool // 处理收到的消息, 这里需要告知RabbitMQ对象消息是返回Ack，
	QueueName() string                      // 获取接收者需要监听的队列
	ConsumeName() string                    // 获取消费者名称
	RouterKey() string                      // 这个队列绑定的路由
}

type RabbitMQ struct {
	Conn           *amqp.Connection
	Lock           *sync.RWMutex
	RabbitUrl      string
	err            error
	endConsumeChan chan struct{}
}

//开始创建一个新的rabitmq对象
func NewRabbitMQ(Username string, Password string, Serveraddr string, ServerPort int, Vhost string) (*RabbitMQ, error) {
	RabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", Username, Password, Serveraddr, ServerPort, Vhost)
	conn, err := amqp.Dial(RabbitUrl) //默认10s心跳,编码(us-en)
	if err != nil {
		return nil, err
	}
	rabbitmq := new(RabbitMQ)
	rabbitmq.Conn = conn
	rabbitmq.Lock = new(sync.RWMutex)
	rabbitmq.RabbitUrl = RabbitUrl
	rabbitmq.endConsumeChan = make(chan struct{})
	return rabbitmq, nil
}

// 关闭队列
func (r *RabbitMQ) Close() {
	if r.Conn != nil {
		r.Conn.Close()
		r.Conn = nil
	}
}

// 获取一个可用的Channel，使用完需要Close
// 建议对 Publish 和 Consume 使用单独的连接
func (r *RabbitMQ) ConnAndChannel() (channel *amqp.Channel, err error) {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	channel, err = r.Conn.Channel()
	return
}

// 创建队列
func (r *RabbitMQ) CreateQueue(channel *amqp.Channel, queue string) (err error) {
	_, err = channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return
}

func (r *RabbitMQ) CreateQueueWithArgs(channel *amqp.Channel, queue string, args amqp.Table) (err error) {
	_, err = channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	return
}

// 创建Exchange
// 注意: Errors returned from this method will close the channel
// kind: "direct", "fanout", "topic" and "headers"
func (r *RabbitMQ) CreateExchange(channel *amqp.Channel, name string, kind MQKind) (err error) {
	return channel.ExchangeDeclare(
		name,
		string(kind),
		true,
		false,
		false,
		false,
		nil)
}

func (r *RabbitMQ) CreateExchangeWithArgs(channel *amqp.Channel, name string, kind MQKind, args amqp.Table) (err error) {
	return channel.ExchangeDeclare(
		name,
		string(kind),
		true,
		false,
		false,
		false,
		args)
}

// 队列绑定
// 注意: If the binding could not complete, an error will be returned and the channel will be closed.
func (r *RabbitMQ) QueueBind(channel *amqp.Channel, name, key, exchange string) (err error) {
	return channel.QueueBind(
		name,
		key,
		exchange,
		false,
		nil)
}

func (r *RabbitMQ) QueueBindWithArgs(channel *amqp.Channel, name, key, exchange string, args amqp.Table) (err error) {
	return channel.QueueBind(
		name,
		key,
		exchange,
		false,
		args)
}

// 删除队列
func (r *RabbitMQ) DeleteQueue(channel *amqp.Channel, queue string) (err error) {
	_, err = channel.QueueDelete(
		queue, // name
		false, // IfUnused
		false, // ifEmpty
		true,  // noWait
	)
	return
}

// 推送信息到队列
func (r *RabbitMQ) PublishQueue(channel *amqp.Channel, exchange string, routeKey string, body string) (err error) {
	err = channel.Publish(
		exchange, // exchange
		routeKey, // routing key
		false,    // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	return
}

func (r *RabbitMQ) PublishQueueByte(channel *amqp.Channel, exchange string, routeKey string, body []byte) (err error) {
	err = channel.Publish(
		exchange, // exchange
		routeKey, // routing key
		false,    // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		})
	return
}

// 消费队列信息
func (r *RabbitMQ) ConsumeQueue(channel *amqp.Channel, receiver MQConsume, processCnt int) error {
	err := channel.Qos(
		3,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if err != nil {
		receiver.OnError(err)
		return err
	}

	msgs, err := channel.Consume(
		receiver.QueueName(),   // queue
		receiver.ConsumeName(), // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)

	if err != nil {
		receiver.OnError(err)
		return err
	}

	var wg sync.WaitGroup
	var endConsume bool = false
	for i := 0; i < processCnt; i++ {
		go func(index int) {
			for d := range msgs {
				wg.Add(1)
				if ok := receiver.OnReceive(d.Body, index); ok {
					if err := d.Ack(false); err != nil {
						receiver.OnError(err)
					}
				}
				wg.Done()

				if endConsume {
					return
				}
			}
		}(i)
	}

	<-r.endConsumeChan
	endConsume = true
	if err := channel.Cancel(receiver.ConsumeName(), true); err != nil {
		receiver.OnError(err)
	}

	wg.Wait()

	return nil

	/*
		for {
			select {
			case d := <-msgs:
				consumeMsg = true
				if ok := receiver.OnReceive(d.Body, d.MessageId); ok {
					_ = d.Ack(false)
				}
				consumeMsg = false
				if endConsume {
					return nil
				}

			case <-r.endConsumeChan:
				endConsume = true
				_ = channel.Cancel(receiver.ConsumeName(), true)

				if !consumeMsg {
					return nil
				}
			}
		}
	*/
}

func (r *RabbitMQ) EndConsumeQueue() {
	r.endConsumeChan <- struct{}{}
}

// 重推消息
func (r *RabbitMQ) ReConsume(exchange string, queue string, msg string) error {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	channel, err := r.Conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	merr := r.PublishQueue(channel, exchange, queue, msg)
	if merr != nil {
		return merr
	}

	return nil
}

// 检查当前队列还有多少消息未被消费
func (r *RabbitMQ) GetReadyCount(channel *amqp.Channel, queue string) (int, error) {
	state, err := channel.QueueInspect(queue)
	if err != nil {
		return 0, err
	}
	return state.Messages, nil
}

// 检查当前队列还有多少消息者
func (r *RabbitMQ) GetConsumCount(channel *amqp.Channel, queue string) (int, error) {
	state, err := channel.QueueInspect(queue)
	if err != nil {
		return 0, err
	}
	return state.Consumers, nil
}

// 删除所有未等待确认的消息，成功时，返回清除的个数
func (r *RabbitMQ) QueuePurge(queue string) (int, error) {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	channel, err := r.Conn.Channel()
	if err != nil {
		return 0, err
	}
	defer channel.Close()

	cnt, err := channel.QueuePurge(queue, true)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}
