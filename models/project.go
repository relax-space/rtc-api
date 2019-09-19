package models

import (
	"context"
	"rtc-api/factory"
	"time"

	"github.com/go-xorm/xorm"
)

type Project struct {
	Id         int    `json:"id" xorm:"pk autoincr"`
	Name       string `json:"name" xorm:"unique"` //service + "|" + namespace
	Service    string `json:"service" xorm:"index notnull"`
	Namespace  string `json:"namespace" xorm:"index"`
	TenantName string `json:"tenantName"`

	SubIds  []int      `json:"subIds" xorm:"varchar(255)"` //subIds
	Setting SettingDto `json:"setting" xorm:"json"`

	Children  []*Project `json:"children" xorm:"-"`
	CreatedAt *time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt *time.Time `json:"updatedAt" xorm:"updated"`
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

func (Project) GetByName(ctx context.Context, name string) (bool, *Project, error) {
	project := &Project{}
	has, err := factory.DB(ctx).Where("name=?", name).Get(project)
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
