package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pangpanglabs/goutils/httpreq"
)

type Gitlab struct {
}

type ApiProject struct {
	Id int `json:"id"`
}

func (d Gitlab) RequestFile(projectDto *ProjectDto, folderName, fileName string) (b []byte, err error) {
	urlstr, err := d.getFileUrl(projectDto.IsMulti,
		projectDto.GitShortPath, projectDto.ServiceName, folderName, fileName)
	if err != nil {
		return
	}
	b, err = (File{}).ReadUrl(urlstr, PRIVATETOKEN)
	if err != nil {
		return
	}
	return
}

func (d Gitlab) CheckTestFile(projectDto *ProjectDto) (err error) {
	urlstr, err := d.getFileUrl(projectDto.IsMulti,
		projectDto.GitShortPath, projectDto.ServiceName, "", "config.test.yml")
	if err != nil {
		return
	}
	_, err = (File{}).ReadUrl(urlstr, PRIVATETOKEN)
	if err != nil {
		return
	}
	return
}

func (d Gitlab) getFileUrl(isMulti bool, gitShortPath, serviceName, folderName, fileName string) (urlstr string, err error) {
	id, err := d.getProjectId(gitShortPath)
	if err != nil {
		return
	}
	name := d.getFilePath(isMulti, serviceName, folderName, fileName)
	urlstr = fmt.Sprintf("%v/api/v4/projects/%v/repository/files/%v/raw?ref=%v",
		PREGITHTTPURL, id, name, app_env)
	return
}

func (d Gitlab) getProjectId(gitShortPath string) (projectId int, err error) {
	groupName, projectName := d.getGroupProject(gitShortPath)
	url := fmt.Sprintf("%v/api/v4/groups/%v/projects?search=%v&simple=true",
		PREGITHTTPURL, groupName, projectName)
	var apiResult []ApiProject
	req := httpreq.New(http.MethodGet, url, nil)
	req.Req.Header.Set("PRIVATE-TOKEN", PRIVATETOKEN)
	_, err = req.Call(&apiResult)
	if err != nil {
		return
	}
	if len(apiResult) != 1 {
		return
	}
	projectId = apiResult[0].Id
	return
}

func (d Gitlab) getFilePath(isMulti bool, projectName, folderName, fileName string) (path string) {
	flag := url.QueryEscape("/")
	if isMulti {
		path = projectName + flag
	}

	if len(folderName) != 0 {
		path += folderName + flag
	}
	path += fileName
	return
}

func (d Gitlab) getGroupProject(gitShortPath string) (groupName, projectName string) {

	strs := strings.Split(gitShortPath, "/")
	if len(strs) != 2 {
		return
	}
	groupName = strs[0]
	projectName = strs[1]
	return

}

func (d Gitlab) FileErr(projectDto *ProjectDto, folderName, fileName string, errParam error) (err error) {
	url := fmt.Sprintf("%v/%v/raw/%v/%v", PREGITHTTPURL, projectDto.GitShortPath, app_env,
		d.getFilePathErr(projectDto.IsMulti, projectDto.ServiceName, folderName, fileName))
	return fmt.Errorf("check gitlab file,url:%v,err:%v", url, errParam)
}

func (d Gitlab) getFilePathErr(isMulti bool, projectName, folderName, fileName string) (path string) {
	if isMulti {
		path = projectName + "/"
	}

	if len(folderName) != 0 {
		path += folderName + "/"
	}
	path += fileName
	return
}
