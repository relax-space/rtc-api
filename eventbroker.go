package main

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
)

type EventBroker struct {
}

func (EventBroker) fetchProducer() (projectDto *ProjectDto, err error) {
	projectDto = &ProjectDto{}
	projectDto.GitRaw = fmt.Sprintf("%v/infra/eventbroker/raw/qa", PREGITHTTPURL)
	urlString := fmt.Sprintf("%v/test_info/kafka-producer/project.yml", projectDto.GitRaw)
	b, err := fetchFromgitlab(urlString, PRIVATETOKEN)
	if err != nil {
		return
	}
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	return
}

func (EventBroker) fetchConsumer() (projectDto *ProjectDto, err error) {
	projectDto = &ProjectDto{}
	projectDto.GitRaw = fmt.Sprintf("%v/infra/eventbroker/raw/qa", PREGITHTTPURL)
	urlString := fmt.Sprintf("%v/test_info/kafka-consumer/project.yml", projectDto.GitRaw)
	b, err := fetchFromgitlab(urlString, PRIVATETOKEN)
	if err != nil {
		return
	}
	if err = yaml.Unmarshal(b, &projectDto); err != nil {
		err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	return
}

func (d EventBroker) SetEventBroker(viper *viper.Viper, port string, streamNames map[string]string) (err error) {
	//set producer
	project, err := d.fetchProducer()
	if err != nil {
		return
	}
	Compose{}.setComposeProducer(viper, port, project)

	//set consumer
	p, err := d.fetchConsumer()
	if err != nil {
		return
	}

	for _, streamName := range streamNames {
		d.setConsumerEnv(p.Envs, streamName)
		p.ServiceName = "event-kafka-consumer-" + streamName
		p.Ports = []string{}
		Compose{}.setComposeConsumer(viper, p)
	}

	//fetch event-broker sql
	err = d.fetchSql()
	return
}

func (EventBroker) fetchSql() (err error) {
	gitRaw := fmt.Sprintf("%v/%v/raw/qa", PREGITHTTPURL, "infra/eventbroker")
	urlString := fmt.Sprintf("%v/test_info%v/table.sql", gitRaw, "/kafka-consumer")
	if err = fetchTofile(urlString,
		fmt.Sprintf("%v/%v.sql", TEMP_FILE, EventBroker_Name),
		PRIVATETOKEN); err != nil {
		err = fmt.Errorf("read table.sql error:%v", err)
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
