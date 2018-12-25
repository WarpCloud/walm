package handlers

import (
	"k8s.io/client-go/tools/cache"
	"k8s.io/api/core/v1"
	"walm/pkg/release"
	"github.com/sirupsen/logrus"
	"walm/pkg/kafka"
	"time"
	"encoding/json"
	"walm/pkg/k8s/informer"
	"walm/pkg/release/manager/helm"
)

type releaseConfigHandler struct {
	releaseConfigHandler *cache.ResourceEventHandlerFuncs
	kafkaClient          *kafka.KafkaClient
	helmClient           *helm.HelmClient
	enabled              bool
}

func (handler *releaseConfigHandler) enable() {
	handler.enabled = true
	if handler.releaseConfigHandler == nil {
		handler.releaseConfigHandler = &cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if !handler.enabled {
					return
				}
				handler.handleReleaseConfigEvent(obj, release.CreateOrUpdate)
			},
			UpdateFunc: func(old, cur interface{}) {
				if !handler.enabled {
					return
				}
				svc := cur.(*v1.Service)
				isDummyService, transwarpMetaStr, releaseName, namespace := handler.parseSvc(svc)
				if isDummyService {
					oldSvc := old.(*v1.Service)
					_, oldTranswarpMetaStr, _, _ := handler.parseSvc(oldSvc)
					if transwarpMetaStr != oldTranswarpMetaStr {
						handler.sendReleaseConfigDeltaEventToKafka(namespace, releaseName, transwarpMetaStr, release.CreateOrUpdate)
					}

				}
			},
			DeleteFunc: func(obj interface{}) {
				if !handler.enabled {
					return
				}
				handler.handleReleaseConfigEvent(obj, release.Delete)
			},
		}
		informer.GetDefaultFactory().Factory.Core().V1().Services().Informer().AddEventHandler(*(handler.releaseConfigHandler))
	}
}

func (handler *releaseConfigHandler) disable() {
	handler.enabled = false
}

func (handler *releaseConfigHandler) handleReleaseConfigEvent(obj interface{}, deltaType release.ReleaseConfigDeltaEventType) {
	svc := obj.(*v1.Service)
	isDummyService, transwarpMetaStr, releaseName, namespace := handler.parseSvc(svc)
	if isDummyService {
		handler.sendReleaseConfigDeltaEventToKafka(namespace, releaseName, transwarpMetaStr, deltaType)
	}
}

func (handler *releaseConfigHandler) parseSvc(svc *v1.Service) (isDummyService bool, transwarpMetaStr, releaseName, namespace string) {
	if len(svc.Labels) > 0 && svc.Labels["transwarp.meta"] == "true" {
		isDummyService = true
		releaseName = svc.Labels["release"]
		namespace = svc.Namespace
		if len(svc.Annotations) > 0 {
			transwarpMetaStr = svc.Annotations["transwarp.meta"]
		}
	}
	return
}

func (handler *releaseConfigHandler) sendReleaseConfigDeltaEventToKafka(namespace, releaseName, transwarpMetaStr string, deltaType release.ReleaseConfigDeltaEventType) {
	if !handler.kafkaClient.Enable {
		logrus.Warnf("failed to send release %s config delta event to kafka : kafka is not enabled", releaseName)
		return
	}

	retryTimes := 5
	for {
		err := handler.doSendReleaseConfigDeltaEventToKafka(namespace, releaseName, transwarpMetaStr, deltaType)
		if err != nil && retryTimes > 0 {
			retryTimes --
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			logrus.Errorf("failed to send release %s config delta event to kafka : %s", releaseName, err.Error())
			return
		}
		logrus.Infof("succeed to send release %s config delta event to kafka", releaseName)
		break
	}
}

func (handler *releaseConfigHandler) doSendReleaseConfigDeltaEventToKafka(namespace, releaseName, transwarpMetaStr string, deltaType release.ReleaseConfigDeltaEventType) error {
	releaseConfigDeltaEvent, err := handler.buildReleaseConfigDeltaEvent(namespace, releaseName, transwarpMetaStr, deltaType)
	if err != nil {
		logrus.Errorf("failed to get release config delta event : %s", err.Error())
		return err
	}
	releaseConfigDeltaEventStr, err := json.Marshal(releaseConfigDeltaEvent)
	if err != nil {
		logrus.Errorf("failed to marshal release config delta event: %s", err.Error())
		return err
	}
	err = handler.kafkaClient.SyncSendMessage(kafka.ReleaseConfigTopic, string(releaseConfigDeltaEventStr))
	if err != nil {
		logrus.Errorf("failed to send release config delta event to kafka : %s", err.Error())
		return err
	}
	return nil
}

func (handler *releaseConfigHandler) buildReleaseConfigDeltaEvent(namespace, releaseName, transwarpMetaStr string,
	deltaType release.ReleaseConfigDeltaEventType) (releaseConfigDeltaEvent *release.ReleaseConfigDeltaEvent, err error) {
	releaseConfigDeltaEvent = &release.ReleaseConfigDeltaEvent{
		Type: deltaType,
	}
	switch deltaType {
	case release.Delete:
		releaseConfigDeltaEvent.Data = release.ReleaseConfig{
			InstanceName: releaseName,
			ConfigSets:   []release.ReleaseConfigSet{},
		}
	case release.CreateOrUpdate:
		releaseConfigDeltaEvent.Data, err = handler.getReleaseConfig(namespace, releaseName, transwarpMetaStr)
		if err != nil {
			logrus.Errorf("failed to get release config : %s", err.Error())
			return
		}
	}
	return
}

func (handler *releaseConfigHandler) getReleaseConfig(namespace, releaseName, transwarpMetaStr string) (releaseConfig release.ReleaseConfig, err error) {
	releaseConfig = release.ReleaseConfig{
		InstanceName: releaseName,
		ConfigSets:   []release.ReleaseConfigSet{},
	}

	releaseInfo, err := handler.helmClient.GetRelease(namespace, releaseName)
	if err != nil {
		logrus.Errorf("failed to get release info : %s", err.Error())
		return releaseConfig, err
	}

	releaseConfig.AppName = releaseInfo.ChartName
	releaseConfig.Version = releaseInfo.ChartVersion

	outputConfigSet, err := handler.buildOutputConfigSet(transwarpMetaStr)
	if err != nil {
		logrus.Errorf("failed to build output config from dummy service: %s", err.Error())
		return releaseConfig, err
	}
	releaseConfig.ConfigSets = append(releaseConfig.ConfigSets, outputConfigSet)
	releaseConfig.ConfigSets = append(releaseConfig.ConfigSets, release.ReleaseConfigSet{
		Name:        "input",
		CreatedBy:   "walm",
		ConfigItems: []release.ReleaseConfigItem{{
			Name: "config",
			Value: releaseInfo.ConfigValues,
			Type: "json",
		}},
	})

	return
}

func (handler *releaseConfigHandler) buildOutputConfigSet(transwarpMetaStr string) (release.ReleaseConfigSet, error) {
	releaseConfigSet := release.ReleaseConfigSet{
		Name:        "output",
		CreatedBy:   "walm",
		ConfigItems: []release.ReleaseConfigItem{},
	}

	if len(transwarpMetaStr) == 0 {
		return releaseConfigSet, nil
	}

	dummyServiceConfig := &release.DummyServiceConfig{}
	err := json.Unmarshal([]byte(transwarpMetaStr), dummyServiceConfig)
	if err != nil {
		logrus.Errorf("failed to unmarshal dummy service config string: %s", err.Error())
		return release.ReleaseConfigSet{}, err
	}

	for key, value := range dummyServiceConfig.Provides {
		releaseConfigItem := release.ReleaseConfigItem{
			Name:  "metadata.provides." + key + ".immediateValue",
			Value: value.ImmediateValue,
			Type:  "json",
		}
		releaseConfigSet.ConfigItems = append(releaseConfigSet.ConfigItems, releaseConfigItem)
	}
	return releaseConfigSet, nil
}
