package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pangpanglabs/goutils/httpreq"
)

type Gitlab struct {
	Url          string
	PrivateToken string
}

type ApiProject struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ApiFile struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (d Gitlab) RequestFile(projectDto *ProjectDto, folderName, subFolderName, fileName, appEnv string) (b []byte, err error) {
	urlstr, err := d.GetFileUrl(projectDto.IsMulti,
		projectDto.GitShortPath, projectDto.ServiceName, folderName, subFolderName, fileName, appEnv)
	if err != nil {
		return
	}
	b, err = (File{}).ReadUrl(urlstr, d.getPrivateToken())
	if err != nil {
		return
	}
	return
}

//Get a list of repository files and directories in a project.
//This endpoint can be accessed without authentication if the repository is publicly accessible.
//https://docs.gitlab.com/ce/api/repositories.html#list-repository-tree
func (d Gitlab) GetFiles(gitShortPath, subPath, ref string) (fileNames []string, err error) {
	projectId, err := d.getProjectId(gitShortPath)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%v/api/v4/projects/%v/repository/tree?path=%v&ref=%v&per_page=1000",
		d.getUrl(), projectId, subPath, ref)
	var apiResult []ApiFile
	req := httpreq.New(http.MethodGet, url, nil)
	req.Req.Header.Set("PRIVATE-TOKEN", d.getPrivateToken())
	_, err = req.Call(&apiResult)
	if err != nil {
		return
	}
	fileNames = make([]string, 0)
	//go-api
	for _, v := range apiResult {
		fileNames = append(fileNames, v.Name)
	}
	return
}

func (d Gitlab) GetFolderPath(isEscape, isMulti bool, projectName, folderName, subFolderName string) (path string) {
	flag := "/"
	if isEscape {
		flag = url.QueryEscape(flag)
		folderName = strings.Replace(folderName, "/", flag, -1)
		subFolderName = strings.Replace(subFolderName, "/", flag, -1)
	}
	if len(folderName) != 0 {
		path += folderName
	}
	if isMulti {
		path += flag + projectName
	}
	if len(subFolderName) != 0 {
		path += flag + subFolderName
	}
	return
}

func (d Gitlab) GetFilePath(isEscape, isMulti bool, projectName, folderName, subFolderName, fileName string) (path string) {
	flag := "/"
	if isEscape {
		flag = url.QueryEscape(flag)
	}
	path = d.GetFolderPath(isEscape, isMulti, projectName, folderName, subFolderName)
	path += flag + fileName
	return
}

func (d Gitlab) CheckTestFile(projectDto *ProjectDto) (err error) {
	err = d.checkTestFile(projectDto, "config.yml")
	if err != nil {
		//if `config.yml` not exist,then don't check `config.rtc.yml`
		if err.Error() == "status:404" {
			err = nil
			return
		}
		return
	}
	err = d.checkTestFile(projectDto, "config.rtc.yml")
	if err != nil {
		return
	}
	return
}

func (d Gitlab) FileErr(projectDto *ProjectDto, folderName, subFolderName, fileName, appEnv string, errParam error) (err error) {
	url := fmt.Sprintf("%v/%v/raw/%v/%v", d.getUrl(), projectDto.GitShortPath, appEnv,
		d.GetFilePath(false, projectDto.IsMulti, projectDto.ServiceName, folderName, subFolderName, fileName))
	return fmt.Errorf("check gitlab file,url:%v,err:%v", url, errParam)
}

func (d Gitlab) GetFileUrl(isMulti bool, gitShortPath, serviceName, folderName, subFolderName, fileName, appEnv string) (urlstr string, err error) {
	id, err := d.getProjectId(gitShortPath)
	if err != nil {
		return
	}
	name := d.GetFilePath(true, isMulti, serviceName, folderName, subFolderName, fileName)
	urlstr = fmt.Sprintf("%v/api/v4/projects/%v/repository/files/%v/raw?ref=%v",
		d.getUrl(), id, name, appEnv)
	return
}

func (d Gitlab) checkTestFile(projectDto *ProjectDto, fileName string) (err error) {
	urlstr, err := d.GetFileUrl(projectDto.IsMulti,
		projectDto.GitShortPath, projectDto.ServiceName, projectDto.ExecPath, "", fileName, app_env)
	if err != nil {
		return
	}
	_, err = (File{}).ReadUrl(urlstr, d.getPrivateToken())
	if err != nil {
		return
	}
	return
}

func (d Gitlab) getProjectId(gitShortPath string) (projectId int, err error) {
	groupName, projectName := d.getGroupProject(gitShortPath)
	url := fmt.Sprintf("%v/api/v4/groups/%v/projects?search=%v&simple=true",
		d.getUrl(), groupName, projectName)
	var apiResult []ApiProject
	req := httpreq.New(http.MethodGet, url, nil)
	req.Req.Header.Set("PRIVATE-TOKEN", d.getPrivateToken())
	_, err = req.Call(&apiResult)
	if err != nil {
		return
	}
	//go-api
	for _, v := range apiResult {
		if v.Name == projectName {
			projectId = v.Id
			return
		}
	}
	err = errors.New("projectId has not found")
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

func (d Gitlab) getUrl() string {
	if len(d.Url) == 0 {
		return comboResource.PerGitHttpUrl
	}
	return d.Url
}

func (d Gitlab) getPrivateToken() string {
	if len(d.PrivateToken) == 0 {
		return comboResource.PrivateToken
	}
	return d.PrivateToken
}
