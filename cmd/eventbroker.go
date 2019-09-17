package cmd

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
	if scope == LOCAL.String() {
		return
	}
	err = ProjectInfo{}.WriteUrlSql(p, comboResource.PrivateToken)
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
	p := &ProjectDto{
		ServiceName: EventBroker_Name,
		Registry:    comboResource.Registry + "/" + serviceName + "-" + app_env,
		Envs:        project.Envs,
		Ports:       []string{inPort.EventBroker},
		Entrypoint:  "./kafka-producer",
		DependsOn:   []string{Compose{}.getServiceServer("kafka")},
	}
	Compose{}.appCompose(viper, p)
}

func (EventBroker) setComposeConsumer(viper *viper.Viper, project *ProjectDto, serverName string) {
	d := Compose{}
	p := &ProjectDto{
		ServiceName: serverName,
		Registry:    comboResource.Registry + "/" + serverName + "-" + app_env,
		Envs:        project.Envs,
		Ports:       project.Ports,
		Entrypoint:  "./kafka-consumer",
		DependsOn: []string{d.getServiceServer("kafka"),
			d.getServiceServer("mysql"),
			d.getServiceServer("redis")},
	}
	Compose{}.appCompose(viper, p)
}

func (d EventBroker) ShouldEventBroker(project *ProjectDto) bool {
	if list := d.StreamList(project); len(list) != 0 {
		return true
	}
	return false
}
func (EventBroker) StreamList(project *ProjectDto) (list map[string]string) {
	list = make(map[string]string, 0)
	for _, d := range project.StreamNames {
		list[d] = d
	}
	for _, subProject := range project.SubProjects {
		for _, d := range subProject.StreamNames {
			list[d] = d
		}
	}
	return
}
