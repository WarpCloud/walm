package kafka

import (
	"testing"
	"walm/pkg/setting"
	"fmt"
)

func TestKafkaClient(t *testing.T) {

	InitKafkaClient(&setting.KafkaConfig{
		Enable: true,
		Brokers: []string{"172.26.0.5:9092"},
	})

	err := kafkaClient.SyncSendMessage("project-status", "test")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}

}
