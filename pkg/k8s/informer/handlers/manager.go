package handlers

import (
	"walm/pkg/release/manager/helm"
	"walm/pkg/kafka"
	"github.com/sirupsen/logrus"
)

var handlers []handler

type handler interface {
	enable()
	disable()
}

func EnableHandlers() {
	if handlers == nil {
		handlers = append(handlers, &releaseConfigHandler{
			helmClient: helm.GetDefaultHelmClient(),
			kafkaClient: kafka.GetDefaultKafkaClient(),
		})
	}

	for _, handler := range handlers {
		handler.enable()
	}
	logrus.Info("informer handlers enabled")
}

func DisableHandlers() {
	for _, handler := range handlers {
		handler.disable()
	}
	logrus.Info("informer handlers disabled")
}
