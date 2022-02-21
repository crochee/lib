package async

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
	"go.uber.org/multierr"

	"github.com/crochee/lirity/mq"
)

// Channel is a channel interface to make testing possible.
// It is highly recommended to use *amqp.Channel as the interface implementation.
type Channel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

func NewRabbitmqChannel(opts ...mq.Option) (Channel, error) {
	client, err := mq.New(opts...)
	if err != nil {
		return nil, err
	}
	if client.IsClosed() {
		return nil, errors.New("rabbitmq is connection closed")
	}
	if !client.IsConnected() {
		return nil, errors.New("not connected to rabbitmq")
	}
	var channel *amqp.Channel
	if channel, err = client.Channel(); err != nil {
		return nil, fmt.Errorf("cann't open channel,%w", err)
	}
	return &rabbitmqChannel{client: client, channel: channel}, nil
}

type rabbitmqChannel struct {
	client  *mq.Client
	channel *amqp.Channel
}

// nolint:gocritic
func (r *rabbitmqChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return r.channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (r *rabbitmqChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool,
	args amqp.Table) (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		queue,
		// 用来区分多个消费者
		consumer,
		// 是否自动应答(自动应答确认消息，这里设置为否，在下面手动应答确认)
		autoAck,
		// 是否具有排他性
		exclusive,
		// 如果设置为true，表示不能将同一个connection中发送的消息
		// 传递给同一个connection的消费者
		noLocal,
		// 是否为阻塞
		noWail,
		args,
	)
}

func (r *rabbitmqChannel) Close() error {
	var errs error
	if err := r.channel.Close(); err != nil {
		errs = multierr.Append(errs, err)
	}
	if err := r.client.Close(); err != nil {
		errs = multierr.Append(errs, err)
	}
	return errs
}
