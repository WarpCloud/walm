package kafka

const (
	ReleaseConfigTopic string = "release-config"
)

type Kafka interface {
	SyncSendMessage(topic, message string) error
}
