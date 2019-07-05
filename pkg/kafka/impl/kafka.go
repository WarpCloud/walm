package impl

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/setting"
)

type Kafka struct {
	setting.KafkaConfig
	syncProducer sarama.SyncProducer
}

func (kafkaImpl *Kafka) SyncSendMessage(topic, message string) error {
	if !kafkaImpl.Enable {
		logrus.Warnf("kafka client is not enabled, failed to send message %s to topic %s", message, topic)
		return nil
	}
	_, _, err := kafkaImpl.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	})

	if err != nil {
		logrus.Errorf("failed to send msg %s to topic %s : %s", message, topic, err.Error())
	}

	logrus.Infof("succeed to send msg %s to topic %s", message, topic)
	return err
}
