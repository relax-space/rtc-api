package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phayes/freeport"

	"github.com/ghodss/yaml"
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
	gitRaw := fmt.Sprintf("%v/%v/raw/qa", PREGITHTTPURL, projectDto.GitShortPath)
	urlString := fmt.Sprintf("%v/test_info/project.yml", gitRaw)
	projectDto.GitRaw = gitRaw
	if err = d.getTestInfo(projectDto, urlString); err != nil {
		return
	}

	path := d.getTestInfoPath(projectDto)

	if projectDto.IsMulti {
		urlString := fmt.Sprintf("%v/test_info%v/project.yml", projectDto.GitRaw, path)
		if err = d.getTestInfo(projectDto, urlString); err != nil {
			return
		}
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

	path := d.getTestInfoPath(project)
	urlString := fmt.Sprintf("%v/test_info%v/table.sql", project.GitRaw, path)
	filePath := fmt.Sprintf("%v/%v.sql", TEMP_FILE, project.ServiceName)
	if err = fetchTofile(urlString, filePath, PRIVATETOKEN); err != nil {
		err = fmt.Errorf("download sql error,url:%v,err:%v", urlString, err)
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
		path := d.getTestInfoPath(projectDto)
		urlString := fmt.Sprintf("%v/test_info%v/table.sql", projectDto.GitRaw, path)
		filePath := fmt.Sprintf("%v/%v.sql", TEMP_FILE, projectDto.ServiceName)
		if err = fetchTofile(urlString, filePath, PRIVATETOKEN); err != nil {
			err = fmt.Errorf("download sql error,url:%v,err:%v", urlString, err)
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

func (d Relation) getTestInfo(projectDto *ProjectDto, urlString string) (err error) {
	b, err := fetchFromgitlab(urlString, PRIVATETOKEN)
	if err != nil {
		return
	}
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = fmt.Errorf("parse test_info error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	return
}

func (d Relation) getTestInfoPath(projectDto *ProjectDto) (path string) {
	if projectDto.IsMulti {
		path += "/" + projectDto.ServiceName
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
	registry = strings.Replace(image, QAREGISTRY, REGISTRYNAME, -1)

	i := strings.LastIndex(registry, "-")
	if i > 0 {
		registry = registry[0:i] + "-qa"
	}
	return
}
