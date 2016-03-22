package zipkin

import (
	"math/rand"

	"github.com/apache/thrift/lib/go/thrift"
	"gopkg.in/Shopify/sarama.v1"

	"github.com/go-kit/kit/log"
)

// defaultKafkaTopic sets the standard Kafka topic our Collector will publish
// on. The default topic for zipkin-receiver-kafka is "zipkin", see:
// https://github.com/openzipkin/zipkin/tree/master/zipkin-receiver-kafka
const defaultKafkaTopic = "zipkin"

// KafkaCollector implements Collector by publishing spans to a Kafka
// broker.
type KafkaCollector struct {
	producer     sarama.AsyncProducer
	logger       log.Logger
	topic        string
	shouldSample Sampler
}

// KafkaOption sets a parameter for the KafkaCollector
type KafkaOption func(c *KafkaCollector)

// KafkaLogger sets the logger used to report errors in the collection
// process. By default, a no-op logger is used, i.e. no errors are logged
// anywhere. It's important to set this option.
func KafkaLogger(logger log.Logger) KafkaOption {
	return func(c *KafkaCollector) { c.logger = logger }
}

// KafkaProducer sets the producer used to produce to Kafka.
func KafkaProducer(p sarama.AsyncProducer) KafkaOption {
	return func(c *KafkaCollector) { c.producer = p }
}

// KafkaTopic sets the kafka topic to attach the collector producer on.
func KafkaTopic(t string) KafkaOption {
	return func(c *KafkaCollector) { c.topic = t }
}

// KafkaSampleRate sets the sample rate used to determine if a trace will be
// sent to the collector. By default, the sample rate is 1.0, i.e. all traces
// are sent.
func KafkaSampleRate(sr Sampler) KafkaOption {
	return func(c *KafkaCollector) { c.shouldSample = sr }
}

// NewKafkaCollector returns a new Kafka-backed Collector. addrs should be a
// slice of TCP endpoints of the form "host:port".
func NewKafkaCollector(addrs []string, options ...KafkaOption) (Collector, error) {
	c := &KafkaCollector{
		logger:       log.NewNopLogger(),
		topic:        defaultKafkaTopic,
		shouldSample: SampleRate(1.0, rand.Int63()),
	}

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
	if c.shouldSample(s.traceID) {
		c.producer.Input() <- &sarama.ProducerMessage{
			Topic: c.topic,
			Key:   nil,
			Value: sarama.ByteEncoder(kafkaSerialize(s)),
		}
	}
	return nil
}

// Close implements Collector.
func (c *KafkaCollector) Close() error {
	return c.producer.Close()
}

func kafkaSerialize(s *Span) []byte {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Encode().Write(p); err != nil {
		panic(err)
	}
	return t.Buffer.Bytes()
}
