package rocketmq

import (
	"context"
	"errors"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/apache/rocketmq-client-go/core"
	"github.com/spf13/cast"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/tracing"
	"time"
)

const (
	mqTypeProducer = "producer"
	mqTypeConsumer = "consumer"
)

var (
	noDefinedErr = errors.New("no defined")
)

type RocketMQConfig struct {
	Topic       string `toml:"topic"`
	GroupID     string `toml:"group_id"`
	NameServer  string `toml:"name_server"`
	AccessKey   string `toml:"access_key"`
	SecretKey   string `toml:"secret_key"`
	Type        string `toml:"type"`         //mq角色 producer consumer
	Model       int    `toml:"model"`        //集群模式 1广播 2集群
	HandleModel int    `toml:"handle_model"` //处理模式 producer:1普通 2有序 3事务 consumer:1普通 2有序
}

type Producer struct {
	rocketmq.Producer
	logger log.Logger
	tracer *go2sky.Tracer
	topic  string
}

func (p *Producer) SendMessageSync(ctx context.Context, msg *rocketmq.Message) (*rocketmq.SendResult, error) {
	return p.goTracerAndExec(ctx, msg, func() (result *rocketmq.SendResult, e error) {
		return p.Producer.SendMessageSync(msg)
	})
}

func (p *Producer) SendMessageOrderly(
	ctx context.Context,
	msg *rocketmq.Message,
	selector rocketmq.MessageQueueSelector,
	arg interface{},
	autoRetryTimes int) (*rocketmq.SendResult, error) {

	return p.goTracerAndExec(ctx, msg, func() (result *rocketmq.SendResult, e error) {
		return p.Producer.SendMessageOrderly(msg, selector, arg, autoRetryTimes)
	})
}

func (p *Producer) SendMessageOneway(ctx context.Context, msg *rocketmq.Message) error {
	_, err := p.goTracerAndExec(ctx, msg, func() (result *rocketmq.SendResult, e error) {
		err := p.Producer.SendMessageOneway(msg)
		return nil, err
	})

	return err
}

func (p *Producer) SendMessageOrderlyByShardingKey(ctx context.Context, msg *rocketmq.Message, shardingkey string) (*rocketmq.SendResult, error) {
	return p.goTracerAndExec(ctx, msg, func() (result *rocketmq.SendResult, e error) {
		return p.Producer.SendMessageOrderlyByShardingKey(msg, shardingkey)
	})
}

func (p *Producer) goTracerAndExec(ctx context.Context, msg *rocketmq.Message, f func() (*rocketmq.SendResult, error)) (*rocketmq.SendResult, error) {
	if go2sky.TraceID(ctx) != go2sky.EmptyTraceID && p.tracer != nil {
		span, err := p.tracer.CreateExitSpan(ctx, "send message", "MQ:"+p.topic, func(header string) error {
			msg.Property[tracing.TraceId] = header
			return nil
		})

		if err != nil {
			p.logger.ErrorT(ctx, "mq send init tracer", "err", err.Error())
			return f()
		}

		resp, err := f()

		if err != nil {
			span.Error(time.Now(), err.Error())
		}

		if resp != nil {
			span.Tag(go2sky.TagStatusCode, cast.ToString(resp.Status))
			span.Tag(tracing.TagMQID, resp.MsgId)
		}

		span.SetSpanLayer(common.SpanLayer_MQ)
		span.End()

		return resp, err
	}

	return f()
}

type Consumer struct {
	rocketmq.PushConsumer
	logger      log.Logger
	tracer      *go2sky.Tracer
	topic       string
	serviceName string
}

func (c *Consumer) Subscribe(ctx context.Context, topic, expression string, consumeFunc func(ctx context.Context, msg *rocketmq.MessageExt) rocketmq.ConsumeStatus) error {
	consumeFuncTracer := func(msg *rocketmq.MessageExt) rocketmq.ConsumeStatus {
		return consumeFunc(context.Background(), msg)
	}

	if c.tracer != nil {
		consumeFuncTracer = func(msg *rocketmq.MessageExt) rocketmq.ConsumeStatus {
			span, ctx, err := c.tracer.CreateEntrySpan(context.Background(), c.serviceName+" subscribe message", func() (s string, e error) {
				return msg.GetProperty(tracing.TraceId), nil
			})

			resp := consumeFunc(ctx, msg)
			if err != nil {
				c.logger.ErrorT(ctx, "mq subscribe init tracer", "err", err.Error())
			} else {
				span.Tag(go2sky.TagStatusCode, resp.String())
				span.Tag(tracing.TagMQID, msg.MessageID)
				span.Tag(tracing.TagMQTopic, msg.Topic)
				span.SetSpanLayer(common.SpanLayer_MQ)
				span.End()
			}

			return resp
		}
	}

	return c.PushConsumer.Subscribe(topic, expression, consumeFuncTracer)
}

type Client struct {
	tracer              *go2sky.Tracer
	logger              log.Logger
	producer            map[string]*Producer
	transactionProducer map[string]*rocketmq.ProducerConfig
	consumer            map[string]*Consumer
}

func (c *Client) GetProducerClient(topic string) (*Producer, error) {
	if producer, ok := c.producer[topic]; ok {
		return producer, nil
	}

	return nil, noDefinedErr
}

func (c *Client) GetTransactionProducerConf(topic string) (*rocketmq.ProducerConfig, error) {
	if config, ok := c.transactionProducer[topic]; ok {
		return config, nil
	}

	return nil, noDefinedErr
}

func (c *Client) GetConsumerClient(topic string) (*Consumer, error) {
	if consumer, ok := c.consumer[topic]; ok {
		return consumer, nil
	}

	return nil, noDefinedErr
}

func (c *Client) Close() {
	if c != nil {
		for _, producer := range c.producer {
			producer.Shutdown()
			//fmt.Println("rocketmq close " + producer.topic, err)
		}
	}

}

func NewClient(serviceName string, configs []RocketMQConfig, logger log.Logger, tracer *go2sky.Tracer) (*Client, error) {
	if len(configs) == 0 {
		return nil, errors.New("config empty")
	}

	client := &Client{
		logger:              logger,
		tracer:              tracer,
		producer:            map[string]*Producer{},
		transactionProducer: map[string]*rocketmq.ProducerConfig{},
		consumer:            map[string]*Consumer{},
	}

	for _, config := range configs {
		c := rocketmq.ClientConfig{
			//您在阿里云 RocketMQ 控制台上申请的 GID。
			GroupID: config.GroupID,
			//设置 TCP 协议接入点，从阿里云 RocketMQ 控制台的实例详情页面获取。
			NameServer: config.NameServer,
			Credentials: &rocketmq.SessionCredentials{
				//您在阿里云账号管理控制台中创建的 AccessKeyId，用于身份认证。
				AccessKey: config.AccessKey,
				//您在阿里云账号管理控制台中创建的 AccessKeySecret，用于身份认证。
				SecretKey: config.SecretKey,
				//用户渠道，默认值为：ALIYUN。
				Channel: "ALIYUN",
			},
		}

		switch config.Type {
		case mqTypeProducer:
			pConfig := &rocketmq.ProducerConfig{
				ClientConfig: c,
			}

			pConfig.ProducerModel = rocketmq.ProducerModel(config.HandleModel)
			if config.HandleModel == 3 {
				client.transactionProducer[config.Topic] = pConfig
			} else {
				p, err := rocketmq.NewProducer(pConfig)
				if err != nil {
					logger.Error("rocketmq producer init error", "topic", config.Topic, "err", err.Error())
					break
				}
				if p.Start() == nil {
					client.producer[config.Topic] = &Producer{Producer: p, logger: logger, tracer: tracer, topic: config.Topic}
				}
			}
		case mqTypeConsumer:
			pConfig := &rocketmq.PushConsumerConfig{
				ClientConfig:  c,
				Model:         rocketmq.MessageModel(config.Model),
				ConsumerModel: rocketmq.ConsumerModel(config.HandleModel),
			}
			cs, err := rocketmq.NewPushConsumer(pConfig)
			if err != nil {
				logger.Error("rocketmq consumer init error", "topic", config.Topic, "err", err.Error())
			}

			client.consumer[config.Topic] = &Consumer{PushConsumer: cs, logger: logger, tracer: tracer, topic: config.Topic, serviceName: serviceName}

		default:
			logger.Error("rocketmq init error", "topic", config.Topic, "err", "mq type err ["+config.Type+"]")

		}
	}

	return client, nil
}
