package main

import (
	"github.com/spf13/viper"
)

type EventBroker struct {
}

func (d EventBroker) SetEventBroker(viper *viper.Viper, port string, streamNames map[string]string) (err error) {
	//set producer
	project, err := d.fetchProducer()
	if err != nil {
		return
	}
	d.setComposeProducer(viper, port, project)

	//set consumer
	p, err := d.fetchConsumer()
	if err != nil {
		return
	}

	for _, streamName := range streamNames {
		d.setConsumerEnv(p.Envs, streamName)
		p.Ports = []string{}
		d.setComposeConsumer(viper, p, "event-kafka-consumer-"+streamName)
	}
	err = ProjectInfo{}.WriteUrl(p, PRIVATETOKEN)
	return
}

func (EventBroker) fetchProducer() (projectDto *ProjectDto, err error) {
	projectDto = &ProjectDto{
		GitShortPath: "infra/eventbroker",
		ServiceName:  "kafka-producer",
		IsMulti:      true,
	}
	if err = (ProjectInfo{}).ReadYml(projectDto); err != nil {
		return
	}
	return
}

func (EventBroker) fetchConsumer() (projectDto *ProjectDto, err error) {
	projectDto = &ProjectDto{
		GitShortPath: "infra/eventbroker",
		ServiceName:  "kafka-consumer",
		IsMulti:      true,
	}
	if err = (ProjectInfo{}).ReadYml(projectDto); err != nil {
		return
	}
	return
}

func (EventBroker) setConsumerEnv(evns []string, streamName string) {
	for k, v := range evns {
		if v == "CONSUMER_GROUP_ID=" {
			evns[k] = v + streamName
		}
	}
}

func (EventBroker) setComposeProducer(viper *viper.Viper, port string, project *ProjectDto) {
	serviceName := EventBroker_Name
	compose := &Compose{
		ServiceName: serviceName,
		ImageName:   REGISTRYELAND + "/" + serviceName + "-" + app_env,
		Restart:     "on-failure:10",
		Environment: project.Envs,
		Ports:       []string{port + ":" + inPort.EventBroker},
		DependsOn:   []string{Compose{}.getServiceServer("kafka")},
	}
	compose.setCompose(viper)
}

func (EventBroker) setComposeConsumer(viper *viper.Viper, project *ProjectDto, serverName string) {
	d := Compose{}
	compose := &Compose{
		ServiceName: serverName,
		ImageName:   REGISTRYELAND + "/" + serverName + "-" + app_env,
		Restart:     "on-failure:10",
		Environment: project.Envs,
		Ports:       project.Ports,
		DependsOn: []string{d.getServiceServer("kafka"),
			d.getServiceServer("mysql"), d.getServiceServer("redis")},
	}
	compose.setCompose(viper)
}
