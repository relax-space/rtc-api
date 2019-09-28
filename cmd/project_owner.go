package cmd

type ProjectOwner struct {
}

func (d ProjectOwner) ReLoad(p *Project) error {
	p.Owner.IsKafka = d.ShouldKafka(p)
	p.Owner.IsMysql = d.ShouldDb(p, MYSQL)
	p.Owner.IsSqlServer = d.ShouldDb(p, SQLSERVER)
	p.Owner.IsRedis = d.ShouldDb(p, REDIS)
	p.Owner.Databases = d.Database(p)
	list := d.Database(p)
	p.Owner.DbNames = d.DatabaseList(list)
	d.SetNames(p)
	d.SetDependLoop(p)
	d.SetStreams(p)
	p.Owner.IsStream = d.ShouldStream(p.Owner.StreamNames)
	if p.Owner.IsStream {
		var err error
		p.Owner.EventProducer, err = Project{}.GetEventProducer()
		if err !=nil{
			return err
		}
		p.Owner.EventConsumer, err = Project{}.GetEventConsumer()
		if err !=nil{
			return err
		}
	}
	return nil
}

func (d ProjectOwner) ShouldKafka(p *Project) bool {
	flag := false
	d.kafka([]*Project{p}, &flag)
	return flag
}

func (d ProjectOwner) ShouldDb(p *Project, dbType DateBaseType) bool {
	list := d.Database(p)
	for k := range list {
		if dbType.String() == k {
			return true
		}
	}
	return false
}

func (d ProjectOwner) ShouldStream(streams []string) bool {
	if len(streams) != 0 {
		return true
	}
	return false
}

func (d ProjectOwner) DatabaseList(list map[string][]string) []string {
	dbNames := make([]string, 0)
	for k := range list {
		dbNames = append(dbNames, k)
	}
	return dbNames
}

func (d ProjectOwner) Database(p *Project) (list map[string][]string) {
	list = make(map[string][]string, 0)
	projects := []*Project{p}
	d.database(list, projects)
	for k, v := range list {
		list[k] = Unique(v)
	}
	return
}

func (d ProjectOwner) database(list map[string][]string, projects []*Project) {
	for _, project := range projects {
		for k, v := range project.Setting.Databases {
			if _, ok := list[k]; ok {
				list[k] = append(list[k], v...)
			} else {
				list[k] = v
			}
		}
		if len(project.Children) != 0 {
			d.database(list, project.Children)
		}
	}
	return
}

func (d ProjectOwner) kafka(projects []*Project, isKafka *bool) {
	for _, project := range projects {
		if project.Setting.IsProjectKafka {
			*isKafka = true
			return
		}
		if len(project.Children) != 0 {
			d.kafka(project.Children, isKafka)
		}
	}
	return
}

func (d ProjectOwner) SetNames(p *Project) {
	d.childNames(p, p)
}

func (d ProjectOwner) childNames(p *Project, pLoop *Project) {
	for _, v := range pLoop.Children {
		p.Owner.ChildNames = append(p.Owner.ChildNames, v.Name)
		if len(v.Children) != 0 {
			d.childNames(v, v)
		}
	}
}

func (d ProjectOwner) SetDependLoop(p *Project) {
	d.setDepend(p)
	d.setDependLoop(p, p)
}

func (d ProjectOwner) setDependLoop(p *Project, pLoop *Project) {
	for k, v := range pLoop.Children {
		pLoop.DependsOn = append(pLoop.DependsOn, v.Name)
		d.setDepend(pLoop.Children[k])
		if len(v.Children) != 0 {
			d.setDependLoop(pLoop.Children[k], pLoop.Children[k])
		}
	}
}
func (d ProjectOwner) setDepend(pLoop *Project) {
	if pLoop.Setting.IsProjectKafka {
		pLoop.DependsOn = append(pLoop.DependsOn, "kafka")
	}
	dbNames := d.DatabaseList(pLoop.Setting.Databases)
	pLoop.DependsOn = append(pLoop.DependsOn, dbNames...)
}

func (d ProjectOwner) SetStreams(p *Project) {
	d.setStreams(p, p)
	p.Owner.StreamNames = append(p.Owner.StreamNames, p.Setting.StreamNames...)
	p.Owner.StreamNames = Unique(p.Owner.StreamNames)
}

func (d ProjectOwner) setStreams(p *Project, pLoop *Project) {
	for _, v := range pLoop.Children {
		p.Owner.StreamNames = append(p.Owner.StreamNames, v.Setting.StreamNames...)
		if len(v.Children) != 0 {
			d.setStreams(v, v)
		}
	}
}
