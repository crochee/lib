package async

import (
	"context"
	"fmt"

	"github.com/json-iterator/go"
	"github.com/streadway/amqp"

	"github.com/crochee/lirity/logger"
	"github.com/crochee/lirity/mq"
	"github.com/crochee/lirity/routine"
	"github.com/crochee/lirity/validator"
)

func NewTaskConsumer(ctx context.Context, opts ...func(*ConsumerOption)) *taskConsumer {
	t := &taskConsumer{
		ConsumerOption: ConsumerOption{
			Pool: routine.NewPool(ctx, routine.Recover(func(ctx context.Context, i interface{}) {
				logger.From(ctx).Sugar().Errorf("%v", i)
			})),
			Manager:     NewManager(),
			Marshal:     mq.DefaultMarshal{},
			JSONHandler: jsoniter.ConfigCompatibleWithStandardLibrary,
			ParamPool:   NewParamPool(),
			Validator:   validator.NewValidator(),
		},
	}
	for _, opt := range opts {
		opt(&t.ConsumerOption)
	}
	return t
}

type ConsumerOption struct {
	Pool        *routine.Pool   // goroutine safe run pool
	Manager     ManagerExecutor // manager executor how to run
	Marshal     mq.MarshalAPI   // mq  assemble request or response
	JSONHandler jsoniter.API
	ParamPool   ParamPool // get Param
	Validator   validator.Validator
}

type taskConsumer struct {
	ConsumerOption
}

func (t *taskConsumer) Register(executors ...Executor) error {
	return t.Manager.Register(executors...)
}

func (t *taskConsumer) Subscribe(channel Channel, queueName string) error {
	t.Pool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			deliveries, err := channel.Consume(
				queueName,
				// 用来区分多个消费者
				"consumer."+queueName,
				// 是否自动应答(自动应答确认消息，这里设置为否，在下面手动应答确认)
				false,
				// 是否具有排他性
				false,
				// 如果设置为true，表示不能将同一个connection中发送的消息
				// 传递给同一个connection的消费者
				false,
				// 是否为阻塞
				false,
				nil,
			)
			if err != nil {
				logger.From(ctx).Error(err.Error())
				fmt.Println(err)
				continue
			}
			t.handleMessage(ctx, deliveries)
		}
	})
	t.Pool.Wait()
	return nil
}

func (t *taskConsumer) handleMessage(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case v := <-deliveries:
			t.Pool.Go(func(ctx context.Context) {
				if err := t.handle(ctx, v); err != nil {
					logger.From(ctx).Error(err.Error())
				}
			})
		}
	}
}

// nolint:gocritic
func (t *taskConsumer) handle(ctx context.Context, d amqp.Delivery) error {
	msgStruct, err := t.Marshal.Unmarshal(&d)
	if err != nil {
		logger.From(ctx).Error(err.Error())
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil { // nolint:gocritic
			return err
		}
		return nil
	}
	logger.From(ctx).Sugar().Infof("consume uuid %s body:%s", msgStruct.UUID, msgStruct.Payload)
	param := t.ParamPool.Get()
	if err = t.JSONHandler.Unmarshal(msgStruct.Payload, param); err != nil {
		logger.From(ctx).Error(err.Error())
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil { // nolint:gocritic
			return err
		}
		return nil
	}
	if err = t.Validator.ValidateStruct(param); err != nil {
		logger.From(ctx).Error(err.Error())
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil { // nolint:gocritic
			return err
		}
		return nil
	}
	err = t.Manager.Run(ctx, param)
	t.ParamPool.Put(param)
	if err != nil {
		logger.From(ctx).Error(err.Error())
		// 当requeue为true时，将该消息排队，以在另一个通道上传递给使用者。
		// 当requeue为false或服务器无法将该消息排队时，它将被丢弃。
		if err = d.Reject(false); err != nil {
			return err
		}
		return nil
	}
	// 手动确认收到本条消息, true表示回复当前信道所有未回复的ack，用于批量确认。
	// false表示回复当前条目
	return d.Ack(false)
}
