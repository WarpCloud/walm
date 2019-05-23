package kafka

import (
	"WarpCloud/walm/pkg/setting"
	"github.com/Shopify/sarama"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
	"github.com/sirupsen/logrus"
)


const (
	ReleaseConfigTopic string = "release-config"
)

var kafkaClient *KafkaClient

func GetDefaultKafkaClient() *KafkaClient {
	if kafkaClient == nil {
		kafkaClient = &KafkaClient{
			KafkaConfig: *setting.Config.KafkaConfig,
		}

		if kafkaClient.Enable {
			config := sarama.NewConfig()
			config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
			config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
			config.Producer.Return.Successes = true
			tlsConfig := kafkaClient.createTlsConfiguration()
			if tlsConfig != nil {
				config.Net.TLS.Config = tlsConfig
				config.Net.TLS.Enable = true
			}

			// On the broker side, you may want to change the following settings to get
			// stronger consistency guarantees:
			// - For your broker, set `unclean.leader.election.enable` to false
			// - For the topic, you could increase `min.insync.replicas`.

			syncProducer, err := sarama.NewSyncProducer(kafkaClient.Brokers, config)
			if err != nil {
				logrus.Fatalln("Failed to start Sarama producer:", err)
			}
			kafkaClient.syncProducer = syncProducer
		} else {
			logrus.Warn("kafka client is not enabled")
		}
	}
	return kafkaClient
}

type KafkaClient struct {
	setting.KafkaConfig
	syncProducer sarama.SyncProducer
}

type NotEnableError struct{}
func (err NotEnableError) Error() string {
	return "kafka client is not enabled"
}

func IsNotEnableError(err error) bool {
	_, ok := err.(NotEnableError)
	return ok
}

func (client *KafkaClient) SyncSendMessage(topic, message string) error {
	if !client.Enable {
		logrus.Warnf("kafka client is not enabled, failed to send message %s to topic %s", message, topic)
		return NotEnableError{}
	}
	_, _, err := client.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	})

	if err != nil {
		logrus.Errorf("failed to send msg %s to topic %s : %s", message, topic, err.Error())
	}

	return err
}

func (client *KafkaClient) createTlsConfiguration() (t *tls.Config) {
	if client.CertFile != "" && client.KeyFile != "" && client.CaFile != "" {
		cert, err := tls.LoadX509KeyPair(client.CertFile, client.KeyFile)
		if err != nil {
			logrus.Fatal(err)
		}

		caCert, err := ioutil.ReadFile(client.CaFile)
		if err != nil {
			logrus.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: client.VerifySsl,
		}
	}
	// will be nil by default if nothing is provided
	return t
}
