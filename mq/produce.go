// Author: crochee
// Date: 2021/9/6

// Package mq
package mq

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/streadway/amqp"
)

// NewProducer return message.Publisher
func NewProducer(mq *Client, opts ...func(*Option)) message.Publisher {
	option := Option{
		Marshal: DefaultMarshal{},
		QueueName: func(topic string) string {
			return topic
		},
	}
	for _, opt := range opts {
		opt(&option)
	}
	return &Producer{
		client:     mq,
		exchange:   option.Exchange,
		routingKey: option.RoutingKey,
		marshal:    option.Marshal,
		queueName:  option.QueueName,
	}
}

type Producer struct {
	client     *Client
	exchange   string
	routingKey string
	marshal    MarshalAPI
	queueName  func(string) string
	wg         sync.WaitGroup
}

func (p *Producer) Publish(topic string, messages ...*message.Message) (err error) {
	if p.client.IsClosed() {
		err = errors.New("AMQP is connection closed")
		return
	}
	if !p.client.IsConnected() {
		err = errors.New("not connected to AMQP")
		return
	}
	p.wg.Add(1)
	defer p.wg.Done()
	// 申请队列,如果不存在会自动创建，存在跳过创建，保证队列存在，消息能发送到队列中
	var channel *amqp.Channel
	if channel, err = p.client.Channel(); err != nil {
		return fmt.Errorf("cann't open channel,%w", err)
	}
	if err = channel.Tx(); err != nil {
		return fmt.Errorf("cann't start transaction,%w", err)
	}
	defer func() {
		if err != nil {
			err = channel.TxRollback()
		}
		_ = channel.Close()
	}()
	if _, err = channel.QueueDeclare(
		p.queueName(topic),
		// 控制队列是否为持久的，当mq重启的时候不会丢失队列
		true,
		// 是否为自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞
		false,
		// 额外属性
		nil,
	); err != nil {
		return
	}
	for _, m := range messages {
		if err = p.PublishMessage(channel, m); err != nil {
			return
		}
	}
	// commit
	if err == nil {
		err = channel.TxCommit()
	}
	return
}

func (p *Producer) PublishMessage(channel *amqp.Channel, msg *message.Message) error {
	amqpMsg, err := p.marshal.Marshal(msg)
	if err != nil {
		return fmt.Errorf("cann't marshal message,%w", err)
	}
	// 发送消息到队列中
	if err = channel.Publish(
		p.exchange,
		p.routingKey,
		// 如果为true，根据exchange类型和routekey类型，如果无法找到符合条件的队列，name会把发送的信息返回给发送者
		false,
		// 如果为true，当exchange发送到消息队列后发现队列上没有绑定的消费者,则会将消息返还给发送者
		false,
		// 发送信息
		amqpMsg,
	); err != nil {
		return fmt.Errorf("cannot publish msg,%w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	p.wg.Wait()
	return p.client.Close()
}

type NoopPublisher struct {
}

func (NoopPublisher) Publish(string, ...*message.Message) error {
	return nil
}

func (NoopPublisher) Close() error {
	return nil
}
