package models

import (
	"context"
	"fmt"
	"rtc-api/factory"
	"time"

	"github.com/go-xorm/xorm"
)

type Project struct {
	Id         int    `json:"id" xorm:"pk autoincr"`
	Name       string `json:"name" xorm:"unique"` //service + "-" + tenantName + "-" + namespace
	Service    string `json:"service" xorm:"index notnull"`
	Namespace  string `json:"namespace" xorm:"index"`
	TenantName string `json:"tenantName"`

	SubIds  []int      `json:"subIds" xorm:"varchar(255)"` //subIds
	Setting SettingDto `json:"setting" xorm:"json"`

	CreatedAt *time.Time      `json:"createdAt" xorm:"created"`
	UpdatedAt *time.Time      `json:"updatedAt" xorm:"updated"`
	Children  []*Project      `json:"children" xorm:"-"`
	DependsOn []string        `json:"dependsOn" xorm:"-"`
	Owner     ProjectOwnerDto `json:"owner" xorm:"-"`
}
type ProjectSimple struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ProjectOwnerDto struct {
	IsKafka     bool
	IsMysql     bool
	IsSqlServer bool
	IsRedis     bool
	IsStream    bool

	DbTypes       []string
	ChildNames    []string
	StreamNames   []string
	EventProducer *Project
	EventConsumer *Project
	Databases     map[string][]DatabaseDto
	ImageAccounts []ImageAccount
}
type DatabaseDto struct {
	TenantName string
	Namespace  string
	DbName     string
}

type SettingDto struct {
	Image          string              `json:"image" xorm:"notnull"`
	Envs           []string            `json:"envs"`
	IsProjectKafka bool                `json:"isProjectKafka"`
	Ports          []string            `json:"ports"`
	Databases      map[string][]string `json:"databases"`
	StreamNames    []string            `json:"streamNames"`
}

func (d *Project) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(d)
}

func (d *Project) Update(ctx context.Context, id int) (int64, error) {
	return factory.DB(ctx).Where("id=?", id).Update(d)
}
func (d Project) Delete(ctx context.Context, id int) (int64, error) {

	ids, err := d.GetParentIds(ctx, id)
	if err != nil {
		return 0, err
	}
	if len(ids) != 0 {
		return 0, fmt.Errorf("Deletion failed, the following microservices depend on it,ids:%v", ids)
	}
	return factory.DB(ctx).Where("id=?", id).Delete(&d)
}

func (Project) GetParentIds(ctx context.Context, id int) ([]int, error) {
	var results []struct {
		Id     int
		SubIds []int
	}
	err := factory.DB(ctx).Table("project").Cols("id", "sub_ids").Find(&results)
	if err != nil {
		return []int{}, err
	}
	ids := make([]int, 0)
	for _, r := range results {
		if ContainInt(r.SubIds, id) {
			ids = append(ids, r.Id)
		}
	}
	return ids, nil
}

func (Project) GetByName(ctx context.Context, name string) (bool, *Project, error) {
	project := &Project{}
	has, err := factory.DB(ctx).Where("name=?", name).Get(project)
	return has, project, err
}
func (Project) GetById(ctx context.Context, id int) (bool, *Project, error) {
	project := &Project{}
	has, err := factory.DB(ctx).Where("id=?", id).Get(project)
	return has, project, err
}

func (Project) GetByIds(ctx context.Context, ids []int) ([]*Project, error) {
	var projects []*Project
	err := factory.DB(ctx).In("id", ids).Find(&projects)
	return projects, err
}

func (Project) GetAllReal(ctx context.Context) ([]*Project, error) {
	var items []*Project
	err := factory.DB(ctx).Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (Project) GetAllSimple(ctx context.Context) ([]ProjectSimple, error) {
	var items []ProjectSimple
	err := factory.DB(ctx).Table("project").Cols("id", "name").Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (Project) GetAll(ctx context.Context, sortby, order []string, offset, limit int) (int64, []*Project, error) {
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx)
		if len(sortby) == 0 && len(order) == 0 {
			sortby = []string{"id"}
			order = []string{"desc"}
		}
		if err := setSortOrder(q, sortby, order); err != nil {
			factory.Logger(ctx).Error(err)
		}
		return q
	}
	var items []*Project
	totalCount, err := queryBuilder().Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (Project) SetName(name, nsName string) string {
	pName := name
	if len(nsName) != 0 {
		pName += NsServiceSplit + nsName
	}
	return pName
}
