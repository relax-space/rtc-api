package eventconsume

import (
	"context"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/kafka"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type EventConsumer struct {
	brokers []string
	topic   string
	groupId string

	filters []Filter
}

type HandlerFunc func(ConsumeContext) error

type ConsumeContext interface {
	Bind(v interface{}) error
	Context() context.Context
	Status() string
}

func NewEventConsumer(groupId string, brokers []string, topic string, filters []Filter) *EventConsumer {
	c := EventConsumer{
		brokers: brokers,
		topic:   topic,
		groupId: groupId,
		filters: filters,
	}

	return &c
}

func (c *EventConsumer) Handle(f HandlerFunc) error {
	consumer, err := kafka.NewConsumerGroup(c.groupId, c.brokers, c.topic)
	if err != nil {
		return err
	}

	messages, err := consumer.Messages()
	if err != nil {
		return err
	}

	go func() {
		for m := range messages {
			status := jsoniter.Get(m.Value, "status").ToString()
			logEntry := logrus.WithFields(logrus.Fields{
				"offset":    m.Offset,
				"partition": m.Partition,
				"topic":     m.Topic,
				"status":    status,
			})

			handler := func(ctx context.Context) error {
				behaviorlogContext := behaviorlog.NewNopContext()
				behaviorlogContext.AuthToken = jsoniter.Get(m.Value, "authToken").ToString()
				behaviorlogContext.RequestID = jsoniter.Get(m.Value, "requestId").ToString()
				behaviorlogContext.Path = m.Topic
				behaviorlogContext.Uri = m.Topic
				behaviorlogContext.WithBizAttrs(map[string]interface{}{
					"offset":    m.Offset,
					"partition": m.Partition,
					"topic":     m.Topic,
					"status":    status,
				})

				ctx = context.WithValue(ctx, behaviorlog.LogContextName, behaviorlogContext)

				c := consumeContext{
					value:  m.Value,
					ctx:    ctx,
					status: status,
				}
				return f(c)
			}

			for i := range c.filters {
				handler = c.filters[i](handler)
			}

			if err := handler(context.Background()); err != nil {
				logEntry.WithError(err).Error("Fail to consume event")
				continue
			}

			logEntry.Info("Success to consume event")
		}
	}()

	return nil
}

type consumeContext struct {
	value  []byte
	ctx    context.Context
	status string
}

func (c consumeContext) Bind(v interface{}) error {
	jsoniter.Get(c.value, "payload").ToVal(v)
	return nil
}

func (c consumeContext) Context() context.Context {
	return c.ctx
}

func (c consumeContext) Status() string {
	return c.status
}
