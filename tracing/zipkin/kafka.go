package zipkin

import (
	"github.com/apache/thrift/lib/go/thrift"
	"gopkg.in/Shopify/sarama.v1"

	"github.com/go-kit/kit/log"
)

// KafkaTopic sets the Kafka topic our Collector will publish on. The
// default topic for zipkin-receiver-kafka is "zipkin", see:
// https://github.com/openzipkin/zipkin/tree/master/zipkin-receiver-kafka
var KafkaTopic = "zipkin"

// KafkaCollector implements Collector by forwarding spans to a Kafka
// service.
type KafkaCollector struct {
	producer sarama.AsyncProducer
	logger   log.Logger
}

// KafkaOption sets a parameter for the KafkaCollector
type KafkaOption func(s *KafkaCollector)

// KafkaLogger sets the logger used to report errors in the collection
// process. By default, a no-op logger is used, i.e. no errors are logged
// anywhere. It's important to set this option.
func KafkaLogger(logger log.Logger) KafkaOption {
	return func(k *KafkaCollector) { k.logger = logger }
}

// KafkaProducer sets the producer used to produce to Kafka.
func KafkaProducer(p sarama.AsyncProducer) KafkaOption {
	return func(c *KafkaCollector) { c.producer = p }
}

// NewKafkaCollector returns a new Kafka-backed Collector. addrs should be a
// slice of TCP endpoints of the form "host:port".
func NewKafkaCollector(addrs []string, options ...KafkaOption) (Collector, error) {
	c := &KafkaCollector{}
	for _, option := range options {
		option(c)
	}
	if c.producer == nil {
		p, err := sarama.NewAsyncProducer(addrs, nil)
		if err != nil {
			return nil, err
		}
		c.producer = p
	}
	if c.logger == nil {
		c.logger = log.NewNopLogger()
	}

	go c.logErrors()
	return c, nil
}

func (c *KafkaCollector) logErrors() {
	for pe := range c.producer.Errors() {
		c.logger.Log("msg", pe.Msg, "err", pe.Err, "result", "failed to produce msg")
	}
}

// Collect implements Collector.
func (c *KafkaCollector) Collect(s *Span) error {
	c.producer.Input() <- &sarama.ProducerMessage{
		Topic: KafkaTopic,
		Key:   nil,
		Value: sarama.ByteEncoder(byteSerialize(s)),
	}
	return nil
}

// Close implements Collector.
func (c *KafkaCollector) Close() error {
	return c.producer.Close()
}

func byteSerialize(s *Span) []byte {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Encode().Write(p); err != nil {
		panic(err)
	}
	return t.Buffer.Bytes()
}
