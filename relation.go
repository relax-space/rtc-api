package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pangpanglabs/goutils/httpreq"
)

type Relation struct {
	Service         string              `json:"service"`
	Container       string              `json:"container"`
	Image           string              `json:"image"`
	GitlabShortName string              `json:"gitlabShortName"`
	Children        map[string]Relation `json:"children"`
}

type ApiResult struct {
	Success bool     `json:"success"`
	Result  Relation `json:"result"`
	Error   ApiError `json:"error"`
}

type ApiError struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

//https://gateway.p2shop.com.cn/mingbai-api/service_groups/docker?name=OrderShipping
func (d Relation) FetchRalation(serviceName string) (project *ProjectDto, err error) {
	url := fmt.Sprintf("%v/mingbai-api/service_groups/docker?name=%v", P2SHOPHOST, serviceName)
	var apiResult ApiResult
	_, err = httpreq.New(http.MethodGet, url, nil).Call(&apiResult)
	if err != nil {
		return
	}
	if apiResult.Success == false {
		err = fmt.Errorf("no data from mingbao api by serviceName:%v.", serviceName)
		return
	}
	r := apiResult.Result

	project = &ProjectDto{
		ServiceName:  r.Service,
		GitShortPath: r.GitlabShortName,
		Registry:     d.getRegistry(r.Image),
	}
	return
}

func (Relation) getRegistry(image string) (registry string) {
	registry = strings.Replace(image, QAREGISTRY, REGISTRYNAME, -1)

	i := strings.LastIndex(registry, "-")
	registry = registry[0:i] + "-qa"
	return
}
