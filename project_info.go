package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
)

type ProjectInfo struct {
}

func (d ProjectInfo) WriteSql(project *ProjectDto) (err error) {

	if err = d.initDatabase(project); err != nil {
		return
	}
	projects := []*ProjectDto{
		project,
	}
	if err = d.writeSubSql(projects); err != nil {
		return
	}
	return
}

const createMysql = "CREATE DATABASE IF NOT EXISTS `$name`;\n"
const createSqlserver = "CREATE DATABASE $name;\n"

func (d ProjectInfo) initDatabase(project *ProjectDto) (err error) {
	list := Database{}.All(project)
	//1.create folder database
	path := "./" + TEMP_FILE + "/database"
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return
	}
	//2. create subfolder like mysql or sqlserver ...
	//2.1. create file init.sql
	//2.2. write sql content
	var mysqlScript string
	var sqlserverScript string
	for k, v := range list {
		if k == MYSQL.String() {
			for _, name := range v {
				mysqlScript += strings.Replace(createMysql, "$name", name, -1)
			}
		}
		if k == SQLSERVER.String() {
			for _, name := range v {
				sqlserverScript += strings.Replace(createSqlserver, "$name", name, -1)
			}
		}
	}
	if len(mysqlScript) != 0 {
		path += "/mysql"
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return
		}
		path += "/init.sql"
		if err = (File{}).WriteString(path, mysqlScript); err != nil {
			return
		}
	}
	path = "./" + TEMP_FILE + "/database"
	if len(sqlserverScript) != 0 {
		path += "/sqlserver"
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return
		}
		path += "/init.sql"
		if err = (File{}).WriteString(path, sqlserverScript); err != nil {
			return
		}
	}
	return
}

func (ProjectInfo) WriteYml(serviceName, fileName, ymlStr string) (err error) {
	path := fmt.Sprintf("%v/%v", TEMP_FILE, serviceName)
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return
	}
	path = fmt.Sprintf("%v/%v", path, fileName)
	if err = (File{}).CreateSafe(path); err != nil {
		return
	}
	if (File{}).WriteString(path, ymlStr); err != nil {
		err = fmt.Errorf("write to %v error:%v", path, err)
		return
	}
	return
}

func (d ProjectInfo) WriteUrl(projectDto *ProjectDto, privateToken string) (err error) {

	dbNames := Database{}.GetDbNamesForData(projectDto)

	for _, v := range dbNames {
		localDbFolderPath := fmt.Sprintf("%v/%v/%v", TEMP_FILE, projectDto.ServiceName, v)
		localDbPath := fmt.Sprintf("%v/table.sql", localDbFolderPath)

		if err = (File{}).DeleteRegex(localDbPath); err != nil {
			return
		}

		urlstr, errd := Gitlab{}.getFileUrl(projectDto.IsMulti,
			projectDto.GitShortPath, projectDto.ServiceName, TEST_INFO, v, "table.sql")

		if errd != nil {
			err = Gitlab{}.FileErr(projectDto, TEST_INFO, v, "table.sql", errd)
			return
		}
		if err = os.MkdirAll(localDbFolderPath, os.ModePerm); err != nil {
			return
		}
		if err = (File{}).CreateEmpty(localDbPath); err != nil {
			return
		}
		if err = (File{}).WriteUrl(urlstr, localDbPath, PRIVATETOKEN); err != nil {
			err = Gitlab{}.FileErr(projectDto, TEST_INFO, v, "table.sql", err)
			return
		}
	}
	return
}

func (d ProjectInfo) ReadYml(projectDto *ProjectDto) (err error) {
	if scope == LOCAL.String() {
		if err = d.readYmlLocal(projectDto); err != nil {
			return
		}
	} else {
		if err = d.readYmlRemote(projectDto); err != nil {
			return
		}
		err = Gitlab{}.CheckTestFile(projectDto)
		if err != nil {
			err = Gitlab{}.FileErr(projectDto, projectDto.ExecPath, "", "config.test.yml", err)
			return
		}
	}
	if projectDto.IsMulti && len(projectDto.Envs) == 0 {
		d.ReadYml(projectDto)
	}
	return
}

func (d ProjectInfo) ShouldKafka(project *ProjectDto) (isKafka bool) {
	if d.ShouldEventBroker(project) {
		return true
	}
	if project.IsProjectKafka {
		isKafka = true
		return
	}
	for _, subProject := range project.SubProjects {
		if subProject.IsProjectKafka {
			isKafka = true
			break
		}
	}
	return
}

func (d ProjectInfo) ShouldEventBroker(project *ProjectDto) bool {
	if list := d.StreamList(project); len(list) != 0 {
		return true
	}
	return false
}
func (ProjectInfo) StreamList(project *ProjectDto) (list map[string]string) {
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

// private method============================

func (d ProjectInfo) writeSubSql(projects []*ProjectDto) (err error) {
	for _, projectDto := range projects {
		if len(projectDto.ServiceName) == 0 {
			continue
		}
		if err = d.WriteUrl(projectDto, PRIVATETOKEN); err != nil {
			return
		}
		if len(projectDto.SubProjects) != 0 {
			if err = d.writeSubSql(projectDto.SubProjects); err != nil {
				return
			}
		}
	}
	return
}

func (d ProjectInfo) shouldWriteYml(projectDto *ProjectDto) bool {
	if projectDto.IsMulti == false || (projectDto.IsMulti && len(projectDto.Envs) != 0) {
		return true
	}
	return false
}

func (d ProjectInfo) readYmlLocal(projectDto *ProjectDto) (err error) {
	path := fmt.Sprintf("%v/%v", TEMP_FILE, projectDto.ServiceName)
	v := viper.New()
	v.SetConfigName(YMLNAMEPROJEC)
	v.AddConfigPath(path)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error config file: %s", err)
	}
	if err := v.Unmarshal(projectDto); err != nil {
		return fmt.Errorf("Fatal error config file: %s", err)
	}
	return
}
func (d ProjectInfo) readYmlRemote(projectDto *ProjectDto) (err error) {
	b, err := Gitlab{}.RequestFile(projectDto, "test_info", "", "project.yml")
	if err != nil {
		err = Gitlab{}.FileErr(projectDto, "test_info", "", "project.yml", err)
		return
	}
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = Gitlab{}.FileErr(projectDto, "test_info", "", "project.yml", err)
		return
	}
	if d.shouldWriteYml(projectDto) {
		ProjectInfo{}.WriteYml(projectDto.ServiceName, "project.yml", string(b))
	}
	return
}
