package async

import (
	"context"
	"fmt"

	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/json-iterator/go"
	"github.com/streadway/amqp"

	"github.com/crochee/lirity/mq"
	"github.com/crochee/lirity/validator"
)

type ProducerOption struct {
	Marshal     mq.MarshalAPI
	Exchange    string
	JSONHandler jsoniter.API
	ParamPool   ParamPool
	Validator   validator.Validator
}

type TaskProducer struct {
	ProducerOption
	wg sync.WaitGroup
}

// NewProducer return message.Publisher
func NewTaskProducer(opts ...func(*ProducerOption)) *TaskProducer {
	t := &TaskProducer{
		ProducerOption: ProducerOption{
			Marshal:     mq.DefaultMarshal{},
			Exchange:    "dcs.api.async",
			JSONHandler: jsoniter.ConfigCompatibleWithStandardLibrary,
			ParamPool:   NewParamPool(),
			Validator:   validator.NewValidator(),
		},
	}
	for _, opt := range opts {
		opt(&t.ProducerOption)
	}
	return t
}

func (t *TaskProducer) Publish(ctx context.Context, channel Channel, routingKey string, param *Param) error {
	t.wg.Add(1)
	defer t.wg.Done()
	if err := t.Validator.ValidateStruct(param); err != nil {
		return err
	}
	data, err := t.JSONHandler.Marshal(param)
	if err != nil {
		return err
	}

	uuid := watermill.NewUUID()

	var amqpMsg amqp.Publishing
	if amqpMsg, err = t.Marshal.Marshal(message.NewMessage(uuid, data)); err != nil {
		return fmt.Errorf("cann't marshal message,%w", err)
	}
	// 发送消息到队列中
	return channel.Publish(
		t.Exchange,
		routingKey,
		// 如果为true，根据exchange类型和routekey类型，如果无法找到符合条件的队列，name会把发送的信息返回给发送者
		false,
		// 如果为true，当exchange发送到消息队列后发现队列上没有绑定的消费者,则会将消息返还给发送者
		false,
		// 发送信息
		amqpMsg,
	)
}

func (t *TaskProducer) GetParam() *Param {
	return t.ParamPool.Get()
}

func (t *TaskProducer) Close() error {
	t.wg.Wait()
	return nil
}
