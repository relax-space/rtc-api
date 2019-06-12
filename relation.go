package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/viper"

	"github.com/pangpanglabs/goutils/httpreq"
)

type Relation struct {
	Service         string               `json:"service"`
	Container       string               `json:"container"`
	Image           string               `json:"image"`
	GitlabShortName string               `json:"gitlabShortName"`
	Children        map[string]*Relation `json:"children"`
}

type ApiResultArray struct {
	Success bool     `json:"success"`
	Result  []string `json:"result"`
	Error   ApiError `json:"error"`
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
func (d Relation) FetchProject(serviceName string, flag *Flag) (project *ProjectDto, err error) {
	relation, err := d.Fetch(serviceName, flag)
	if err != nil {
		return
	}
	project, err = d.setProject(relation)
	return
}

func (d Relation) Fetch(serviceName string, flag *Flag) (relation *Relation, err error) {

	if BoolPointCheck(flag.RelationSource) {
		// from gitlab
		relation, err = d.fetchFromGitlab(serviceName)
		return
	}

	relation, err = d.fetchFromMingbai(serviceName)
	return
}

func (d Relation) FetchAllNames() (names []string, err error) {

	url := fmt.Sprintf("%v/mingbai-api/service_groups/items/names?runtimeEnv=%v", P2SHOPHOST, "")
	var apiResult ApiResultArray
	_, err = httpreq.New(http.MethodGet, url, nil).Call(&apiResult)
	if err != nil {
		return
	}
	if apiResult.Success == false {
		err = fmt.Errorf("no data from mingbai api ,url:%v", url)
		return
	}

	names = apiResult.Result
	return
}

func (d Relation) fetchFromMingbai(serviceName string) (relation *Relation, err error) {
	//strings.ToUpper(app_env)
	url := fmt.Sprintf("%v/mingbai-api/service_groups/docker?name=%v&runtimeEnv=%v", P2SHOPHOST, serviceName, "")
	var apiResult ApiResult
	_, err = httpreq.New(http.MethodGet, url, nil).Call(&apiResult)
	if err != nil {
		return
	}
	if apiResult.Success == false {
		err = fmt.Errorf("no data from mingbai api ,url:%v", url)
		return
	}
	relation = &apiResult.Result
	return
}

func (d Relation) fetchFromGitlab(serviceName string) (relation *Relation, err error) {
	projectDto := &ProjectDto{
		ServiceName:  "rtc-data",
		GitShortPath: "data/rtc-data",
	}
	//this is qa not app_env by xiao.xinmiao
	appEnv := "qa"
	fileName := fmt.Sprintf("config.%v.yml", appEnv)
	b, err := Gitlab{}.RequestFile(projectDto, "", "", fileName, appEnv)
	if err != nil {
		err = Gitlab{}.FileErr(projectDto, "", "", fileName, app_env, err)
		return
	}
	vip := viper.New()
	vip.SetConfigType("yaml")
	if err = vip.ReadConfig(bytes.NewBuffer(b)); err != nil {
		err = Gitlab{}.FileErr(projectDto, "", "", fileName, app_env, err)
		return
	}
	var rs []map[string]*Relation
	if err = vip.Unmarshal(&rs); err != nil {
		return
	}

	has, r := d.getRelation(serviceName, rs)
	if has == false {
		err = Gitlab{}.FileErr(projectDto, "", "", fileName, app_env, fmt.Errorf("service is missing:%v", serviceName))
		return
	}

	err = d.setRelationDetail(r, rs)
	if err != nil {
		err = Gitlab{}.FileErr(projectDto, "", "", fileName, app_env, err)
	}
	relation = r
	return
}

func (d Relation) setRelationDetail(relation *Relation, rs []map[string]*Relation) (err error) {
	if len(relation.Children) != 0 {
		for k := range relation.Children {
			has, r := d.getRelation(k, rs)
			if has == false {
				err = fmt.Errorf("service is missing:%v", k)
				return
			}
			relation.Children[k] = r
			if len(relation.Children[k].Children) != 0 {
				if err = d.setRelationDetail(relation.Children[k], rs); err != nil {
					return
				}
			}
		}
	}
	return
}

func (d Relation) getRelation(serviceName string, rs []map[string]*Relation) (has bool, relation *Relation) {
	for _, m := range rs {
		for k, v := range m {
			if k == serviceName && v != nil {
				has = true
				relation = v
			}
		}
	}
	return
}

func (d Relation) setProject(r *Relation) (project *ProjectDto, err error) {
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
	if err = (ProjectInfo{}).ReadYml(projectDto); err != nil {
		return
	}
	return
}

func (d Relation) setSubProject(relations map[string]*Relation, project *ProjectDto) (err error) {
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
