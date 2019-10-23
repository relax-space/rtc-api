package controllers

import (
	"context"
	"rtc-api/models"
)

type ProjectOwner struct {
}

func (d ProjectOwner) Reload(ctx context.Context, p *models.Project) error {
	p.Owner.IsKafka = d.ShouldKafka(p)
	p.Owner.IsMysql = d.ShouldDb(p, MYSQL)
	p.Owner.IsSqlServer = d.ShouldDb(p, SQLSERVER)
	p.Owner.IsRedis = d.ShouldDb(p, REDIS)
	d.SetStreams(p)
	p.Owner.IsStream = d.ShouldStream(p.Owner.StreamNames)
	if err := d.SetEvent(ctx, p); err != nil {
		return err
	}

	list := d.Database(p)
	p.Owner.Databases = list
	p.Owner.DbTypes = d.DatabaseTypes(list)
	d.SetNames(p)
	d.SetDependLoop(p)
	if err := d.SetImageAccount(ctx, p); err != nil {
		return err
	}
	return nil
}
func (d ProjectOwner) SetImageAccount(ctx context.Context, p *models.Project) error {
	var err error
	p.Owner.ImageAccounts, err = models.ImageAccount{}.GetAll(ctx)
	return err
}

func (d ProjectOwner) SetEvent(ctx context.Context, p *models.Project) error {
	if p.Owner.IsStream {
		var err error
		_, p.Owner.EventProducer, err = models.Project{}.GetByName(ctx, "event-broker-kafka")
		if err != nil {
			return err
		}
		_, p.Owner.EventConsumer, err = models.Project{}.GetByName(ctx, "event-kafka-consumer")
		if err != nil {
			return err
		}
		p.Owner.IsKafka = true
		p.Owner.IsMysql = true
		p.Owner.IsRedis = true

		list := d.Database(p.Owner.EventConsumer)
		if p.Setting.Databases == nil {
			p.Setting.Databases = make(map[string][]string, 0)
		}
		for k, v := range list {
			if _, ok := list[k]; ok {
				p.Setting.Databases[k] = append(p.Setting.Databases[k], d.removeTenant(v)...)
			} else {
				p.Setting.Databases[k] = d.removeTenant(v)
			}
		}
	}
	return nil
}
func (d ProjectOwner) removeTenant(dbDtos []models.DatabaseDto) []string {
	dbNames := make([]string, 0)
	for _, v := range dbDtos {
		dbNames = append(dbNames, v.DbName)
	}
	return dbNames

}
func (d ProjectOwner) ShouldKafka(p *models.Project) bool {
	flag := false
	d.kafka([]*models.Project{p}, &flag)
	return flag
}

func (d ProjectOwner) ShouldDb(p *models.Project, dbType DateBaseType) bool {
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

func (d ProjectOwner) DatabaseTypes(list map[string][]models.DatabaseDto) []string {
	dbTypes := make([]string, 0)
	for k := range list {
		dbTypes = append(dbTypes, k)
	}
	return dbTypes
}

func (d ProjectOwner) DatabaseTypesString(list map[string][]string) []string {
	dbTypes := make([]string, 0)
	for k := range list {
		dbTypes = append(dbTypes, k)
	}
	return dbTypes
}

func (d ProjectOwner) Database(p *models.Project) (list map[string][]models.DatabaseDto) {
	list = make(map[string][]models.DatabaseDto, 0)
	projects := []*models.Project{p}
	d.database(list, projects)
	for k, v := range list {
		list[k] = d.Unique(v)
	}
	return
}

func (d ProjectOwner) Unique(params []models.DatabaseDto) (list []models.DatabaseDto) {
	list = make([]models.DatabaseDto, 0)
	temp := make(map[string]models.DatabaseDto, 0)
	for _, v := range params {
		temp[v.Namespace+v.DbName] = v
	}
	for _, v := range temp {
		list = append(list, v)
	}
	return
}

func (d ProjectOwner) database(list map[string][]models.DatabaseDto, projects []*models.Project) {
	for _, project := range projects {
		for k, v := range project.Setting.Databases {
			if _, ok := list[k]; ok {
				list[k] = append(list[k], d.addTenant(v, project.TenantName, project.Namespace)...)
			} else {
				list[k] = d.addTenant(v, project.TenantName, project.Namespace)
			}
		}
		if len(project.Children) != 0 {
			d.database(list, project.Children)
		}
	}
	return
}
func (d ProjectOwner) addTenant(dbNames []string, tenantName, namespace string) []models.DatabaseDto {
	dbDtos := make([]models.DatabaseDto, 0)
	for _, dbName := range dbNames {
		dbDtos = append(dbDtos, models.DatabaseDto{
			TenantName: tenantName,
			DbName:     dbName,
			Namespace:  namespace,
		})
	}
	return dbDtos
}

func (d ProjectOwner) kafka(projects []*models.Project, isKafka *bool) {
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

func (d ProjectOwner) SetNames(p *models.Project) {
	d.childNames(p, p)
}

func (d ProjectOwner) childNames(p *models.Project, pLoop *models.Project) {
	for _, v := range pLoop.Children {
		p.Owner.ChildNames = append(p.Owner.ChildNames, v.Name)
		if len(v.Children) != 0 {
			d.childNames(v, v)
		}
	}
}

func (d ProjectOwner) SetDependLoop(p *models.Project) {
	d.setDepend(p)
	d.setDependLoop(p, p)
}

func (d ProjectOwner) setDependLoop(p *models.Project, pLoop *models.Project) {
	for k, v := range pLoop.Children {
		pLoop.DependsOn = append(pLoop.DependsOn, v.Name)
		d.setDepend(pLoop.Children[k])
		if len(v.Children) != 0 {
			d.setDependLoop(pLoop.Children[k], pLoop.Children[k])
		}
	}
}
func (d ProjectOwner) setDepend(pLoop *models.Project) {
	if pLoop.Setting.IsProjectKafka {
		pLoop.DependsOn = append(pLoop.DependsOn, "kafka")
	}
	dbTypes := d.DatabaseTypesString(pLoop.Setting.Databases)
	pLoop.DependsOn = append(pLoop.DependsOn, dbTypes...)
}

func (d ProjectOwner) SetStreams(p *models.Project) {
	d.setStreams(p, p)
	p.Owner.StreamNames = append(p.Owner.StreamNames, p.Setting.StreamNames...)
	p.Owner.StreamNames = Unique(p.Owner.StreamNames)
}

func (d ProjectOwner) setStreams(p *models.Project, pLoop *models.Project) {
	for _, v := range pLoop.Children {
		p.Owner.StreamNames = append(p.Owner.StreamNames, v.Setting.StreamNames...)
		if len(v.Children) != 0 {
			d.setStreams(v, v)
		}
	}
}
