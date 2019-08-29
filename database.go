package main

type Database struct {
}

func (d Database) GetDbNamesForData(projectDto *ProjectDto) (dbNames []string) {
	dbNames = make([]string, 0)
	if d.ShouldDb(projectDto, MYSQL) == true {
		dbNames = append(dbNames, MYSQL.String())
	}
	if d.ShouldDb(projectDto, SQLSERVER) == true {
		dbNames = append(dbNames, SQLSERVER.String())
	}
	return
}
func (d Database) ShouldDbLoop(project *ProjectDto, dbType DateBaseType) (flag bool) {

	list := d.All(project, true)
	for k := range list {
		if dbType.String() == k {
			return true
		}
	}
	return false
}

func (d Database) ShouldDb(project *ProjectDto, dbType DateBaseType) (flag bool) {

	list := d.All(project, false)
	for k := range list {
		if dbType.String() == k {
			return true
		}
	}
	return false
}

func (d Database) All(project *ProjectDto, isLoop bool) (list map[string][]string) {
	list = make(map[string][]string, 0)
	projects := []*ProjectDto{project}
	d.all(list, projects, isLoop)

	for k, v := range list {
		list[k] = Unique(v)
	}
	if (EventBroker{}).ShouldEventBroker(project) {
		list[MYSQL.String()] = append(list[MYSQL.String()], "event_broker")
		list[REDIS.String()] = append(list[REDIS.String()], "")
	}

	return
}

func (d Database) all(list map[string][]string, projects []*ProjectDto, isLoop bool) {
	for _, project := range projects {
		for k, v := range project.Databases {
			if _, ok := list[k]; ok {
				list[k] = append(list[k], v...)
			} else {
				list[k] = v
			}
		}
		if isLoop && len(project.SubProjects) != 0 {
			d.all(list, project.SubProjects, isLoop)
		}
	}
	return
}
