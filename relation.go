package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phayes/freeport"

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
	//strings.ToUpper(app_env)
	url := fmt.Sprintf("%v/mingbai-api/service_groups/docker?name=%v&namespace=%v", P2SHOPHOST, serviceName, "")
	var apiResult ApiResult
	_, err = httpreq.New(http.MethodGet, url, nil).Call(&apiResult)
	if err != nil {
		return
	}
	if apiResult.Success == false {
		err = fmt.Errorf("no data from mingbai api ,url:%v", url)
		return
	}

	project, err = d.setProject(apiResult.Result)
	return
}

func (d Relation) setProject(r Relation) (project *ProjectDto, err error) {
	project = &ProjectDto{
		ServiceName:  r.Service,
		GitShortPath: r.GitlabShortName,
		Registry:     d.getRegistry(r.Image),
	}
	if err = d.setProjectDetail(project); err != nil {
		return
	}
	if err = d.setSubProject(r.Children, project); err != nil {
		return
	}
	return
}

func (d Relation) setProjectDetail(projectDto *ProjectDto) (err error) {
	if len(projectDto.ServiceName) == 0 {
		return
	}
	if err = getProjectEnv(projectDto); err != nil {
		return
	}
	d.setPort(projectDto)
	return
}

func (d Relation) setPort(projectDto *ProjectDto) {
	ports, err := freeport.GetFreePorts(len(projectDto.Ports))
	if err != nil {
		err = fmt.Errorf("get free port error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	for i, _ := range projectDto.Ports {
		projectDto.Ports[i] = fmt.Sprintf("%v:%v", ports[i], projectDto.Ports[i])
	}
}

func (d Relation) FetchsqlTofile(project *ProjectDto) (err error) {
	if err = fetchSqlTofile(project, PRIVATETOKEN); err != nil {
		return
	}
	if err = d.fetchSubsqlTofile(project.SubProjects); err != nil {
		return
	}
	return
}

func (d Relation) fetchSubsqlTofile(projects []*ProjectDto) (err error) {
	for _, projectDto := range projects {
		if len(projectDto.ServiceName) == 0 {
			continue
		}
		if err = fetchSqlTofile(projectDto, PRIVATETOKEN); err != nil {
			return
		}
		if len(projectDto.SubProjects) != 0 {
			if err = d.fetchSubsqlTofile(projectDto.SubProjects); err != nil {
				return
			}
		}
	}
	return
}

func (d Relation) setSubProject(relations map[string]Relation, project *ProjectDto) (err error) {
	for _, r := range relations {
		p := &ProjectDto{
			ServiceName:  r.Service,
			GitShortPath: r.GitlabShortName,
			Registry:     d.getRegistry(r.Image),
		}
		if err = d.setProjectDetail(p); err != nil {
			return
		}
		project.SubProjects = append(project.SubProjects, p)
		if len(r.Children) != 0 {
			if err = d.setSubProject(r.Children, p); err != nil {
				return
			}
		}
	}
	return
}

func (Relation) getRegistry(image string) (registry string) {
	registry = strings.Replace(image, REGISTRYQA, REGISTRYELAND, -1)

	i := strings.LastIndex(registry, "-")
	if i > 0 {
		registry = registry[0:i] + "-" + app_env
	}
	return
}
