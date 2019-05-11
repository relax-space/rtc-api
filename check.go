package main

import (
	"os"
	"strings"
)

func shouldLocalConfig() bool {
	if updatedConfig == LOCAL.String() {
		return true
	}
	return false
}

func shouldLocalFetchsql(serviceName string) bool {
	if shouldLocalConfig() && sqlSettings(serviceName) {
		return true
	}
	return false
}

func shouldLocalProjectYml(serviceName string) bool {
	if shouldLocalConfig() && projectYmlSettings(serviceName) {
		return true
	}
	return false
}

func projectYmlSettings(serviceName string) (isLocal bool) {
	//not found file
	if _, err := os.Stat(TEMP_FILE + "/" + serviceName + "/" + YMLNAMEPROJEC + ".yml"); err != nil {
		isLocal = false
		return
	}
	isLocal = true
	return
}

func sqlSettings(serviceName string) (isLocal bool) {
	//not found file
	if _, err := os.Stat(TEMP_FILE + "/" + serviceName + ".sql"); err != nil {
		isLocal = false
		return
	}
	isLocal = true
	return
}

func shouldUpdateCompose(scope string) bool {
	if _, err := os.Stat(YMLNAMEDOCKERCOMPOSE + ".yml"); err != nil {
		return true
	}
	return scope != LOCAL.String()
}

func shouldStartKakfa(project *ProjectDto) (isKafka bool) {
	if shouldStartEventBroker(project) {
		return true
	}
	if project.IsProjectKafka {
		isKafka = true
	} else {
		for _, subProject := range project.SubProjects {
			if subProject.IsProjectKafka {
				isKafka = true
				break
			}
		}
	}
	return
}

func shouldStartMysql(project *ProjectDto) bool {
	if shouldStartEventBroker(project) {
		return true
	}
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == MYSQL.String() {
			return true
		}
	}
	return false
}

func shouldStartRedis(project *ProjectDto) bool {
	if shouldStartEventBroker(project) {
		return true
	}
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == REDIS.String() {
			return true
		}
	}
	return false
}

func shouldStartMongo(project *ProjectDto) bool {
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == MONGO.String() {
			return true
		}
	}
	return false
}

func shouldStartSqlServer(project *ProjectDto) bool {
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == SQLSERVER.String() {
			return true
		}
	}
	return false
}

func shouldStartEventBroker(project *ProjectDto) bool {
	if list := streamList(project); len(list) != 0 {
		return true
	}
	return false
}

func databaseList(project *ProjectDto) (list map[string]string) {
	list = make(map[string]string, 0)
	for _, d := range project.Databases {
		list[d] = d
	}
	for _, subProject := range project.SubProjects {
		for _, d := range subProject.Databases {
			list[d] = d
		}
	}
	return
}

func streamList(project *ProjectDto) (list map[string]string) {
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

func shouldWriteProjectYml(projectDto *ProjectDto) bool {
	if projectDto.IsMulti == false || (projectDto.IsMulti && len(projectDto.Envs) != 0) {
		return true
	}
	return false
}
