package async

import (
	"context"
	"testing"
	"time"

	"github.com/streadway/amqp"

	"github.com/crochee/lirity/mq"
)

// mockChannel is a mock of Channel.
type mockChannel struct {
	deliveries chan amqp.Delivery
}

// Publish runs a test function f and sends resultant message to a channel.
func (ch *mockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return ch.handle(msg)
}

func (ch *mockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	dev := make(chan amqp.Delivery)
	go func() {
		v := <-ch.deliveries
		dev <- v
	}()
	return dev, nil
}

func (ch *mockChannel) Close() error {
	return nil
}

func (ch *mockChannel) handle(msg amqp.Publishing) error {
	ch.deliveries <- amqp.Delivery{
		Acknowledger:    mockAck{},
		Headers:         msg.Headers,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,
		ConsumerTag:     "",
		MessageCount:    0,
		DeliveryTag:     0,
		Redelivered:     false,
		Exchange:        "",
		RoutingKey:      "",
		Body:            msg.Body,
	}
	return nil
}

type mockAck struct {
}

func (m mockAck) Ack(tag uint64, multiple bool) error {
	return nil
}

func (m mockAck) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (m mockAck) Reject(tag uint64, requeue bool) error {
	return nil
}

func TestInteract(t *testing.T) {
	c := &mockChannel{deliveries: make(chan amqp.Delivery, 10)}
	tp := NewTaskProducer()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tc := NewTaskConsumer(ctx)
	if err := tc.Register(test{}, &test1{}, &multiTest{list: []Executor{test{}, &test1{}}}); err != nil {
		t.Fatal(err)
	}
	if err := tp.Publish(context.Background(), c, "", &Param{
		Name: "async.multiTest",
		Data: nil,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tc.Subscribe(c, ""); err != nil {
		t.Fatal(err)
	}
}

func TestProduce(t *testing.T) {
	cp, err := NewRabbitmqChannel(func(option *mq.Option) {
		option.URI = "amqp://admin:1234567@localhost:5672/"
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cp.Close()
	tp := NewTaskProducer()
	for i := 0; i < 20; i++ {
		if err = tp.Publish(context.Background(), cp, "msg.dcs.woden", &Param{
			Name: "async.multiTest",
			Data: nil,
		}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 20; i++ {
		if err = tp.Publish(context.Background(), cp, "msg.dcs.woden", &Param{
			Name: "async.testError",
			Data: nil,
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConsume(t *testing.T) {
	cc, err := NewRabbitmqChannel(func(option *mq.Option) {
		option.URI = "amqp://admin:1234567@localhost:5672/"
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	tc := NewTaskConsumer(context.Background())
	if err = tc.Register(testError{}, test{}, &test1{}, &multiTest{list: []Executor{test{}, &test1{}}}); err != nil {
		t.Fatal(err)
	}
	if err = tc.Subscribe(cc, "msg.dcs.woden"); err != nil {
		t.Fatal(err)
	}
}
