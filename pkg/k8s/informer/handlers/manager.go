package handlers

import (
	"github.com/sirupsen/logrus"
	"walm/pkg/release/v2/config"
)

var handlers []Handler

type Handler interface {
	Start(stopChan <-chan struct{})
}

func StartHandlers(stopChan <-chan struct{}) {
	if handlers == nil {
		handlers = append(handlers, newReleaseConfigHandler())
		handlers = append(handlers, config.NewReleaseConfigController())
	}

	for _, handler := range handlers {
		go handler.Start(stopChan)
	}
	logrus.Info("informer handlers started")
}

