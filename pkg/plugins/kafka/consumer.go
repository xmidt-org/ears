package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	sync.RWMutex
	sarama.ConsumerGroupSession
	wg      sync.WaitGroup
	ready   chan bool
	ctx     context.Context
	cancel  context.CancelFunc
	client  sarama.ConsumerGroup
	topics  []string
	handler func(message *sarama.ConsumerMessage) bool
}

type KafkaConfig struct {
	Brokers             string `yaml:"brokers,omitempty"`
	Username            string `yaml:"username,omitempty"`
	Password            string `yaml:"password,omitempty"`
	CACert              string `yaml:"caCert,omitempty"`
	AccessCert          string `yaml:"accessCert,omitempty"`
	AccessKey           string `yaml:"accessKey,omitempty"`
	Version             string `yaml:"version,omitempty"`
	ChannelBufferSize   int    `yaml:"channelBufferSize,omitempty"`
	ConsumeByPartitions bool
	TLSEnable           bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	consumer.Lock()
	consumer.ConsumerGroupSession = session
	consumer.Unlock()
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		//		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		if consumer.handler(message) {
			session.MarkMessage(message, "")
		}
	}
	return nil
}

// TODO: avoid 2nd start
func (consumer *Consumer) Start(ctx context.Context, groupHandler sarama.ConsumerGroupHandler, handler func(*sarama.ConsumerMessage) bool) {
	consumer.handler = handler
	consumer.wg.Add(1)
	defer consumer.wg.Done()
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		if err := consumer.client.Consume(consumer.ctx, consumer.topics, groupHandler); err != nil {
			//ctx.Logger().Error("op", "Consumer.Start", "error", err)
		}
		// check if context was cancelled, signaling that the consumer should stop
		if nil != consumer.ctx.Err() {
			return
		}
		consumer.ready = make(chan bool)
	}
}

func (consumer *Consumer) Close(ctx context.Context) {
	<-consumer.ready // Await till the consumer has been set up
	consumer.cancel()
	consumer.wg.Wait()
	if err := consumer.client.Close(); err != nil {
		//ctx.Logger().Error("op", "Consumer.Close", "error", err)
	}
}

func NewConsumer(ctx context.Context, conf *KafkaConfig, topics []string, group string, commitInterval int) (*Consumer, error) {
	return NewConsumerWithContext(ctx, context.Background(), conf, topics, group, commitInterval)
}

func NewConsumerWithContext(ctx context.Context, goCtx context.Context, conf *KafkaConfig, topics []string, group string, commitInterval int) (*Consumer, error) {
	config, err := newConsumerConfig(ctx, conf, commitInterval)
	if nil != err {
		return nil, err
	}
	client, err := sarama.NewConsumerGroup(strings.Split(conf.Brokers, ","), group, config)
	if nil != err {
		return nil, err
	}
	cctx, cancel := context.WithCancel(goCtx)
	return &Consumer{
		ready:  make(chan bool),
		client: client,
		topics: topics,
		ctx:    cctx,
		cancel: cancel,
	}, nil
}

func newConsumerConfig(ctx context.Context, conf *KafkaConfig, commitInterval int) (*sarama.Config, error) {
	// init (custom) config, enable errors and notifications
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Interval = time.Duration(commitInterval) * time.Second
	// config.Group.Return.Notifications = true
	// if conf.ConsumeByPartitions {
	// 	config.Group.Mode = cluster.ConsumerModePartitions
	// }
	if err := setConfig(ctx, conf, config); nil != err {
		return nil, err
	}
	return config, nil
}

func setConfig(ctx context.Context, conf *KafkaConfig, config *sarama.Config) error {
	if "" != conf.Version {
		v, err := sarama.ParseKafkaVersion(conf.Version)
		if nil != err {
			return err
		}
		config.Version = v
	}
	if 0 < conf.ChannelBufferSize {
		config.ChannelBufferSize = conf.ChannelBufferSize
	}
	config.Net.TLS.Enable = conf.TLSEnable
	if "" != conf.Username {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.User = conf.Username
		config.Net.SASL.Password = conf.Password
	} else if "" != conf.AccessCert {
		keypair, err := tls.X509KeyPair([]byte(conf.AccessCert), []byte(conf.AccessKey))
		if err != nil {
			return err
		}
		caAuthorityPool := x509.NewCertPool()
		caAuthorityPool.AppendCertsFromPEM([]byte(conf.CACert))
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caAuthorityPool,
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	return nil
}
