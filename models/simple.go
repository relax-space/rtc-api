package models

import (
	"context"
	"rtc-api/factory"
)

type Tenant struct {
	Id   int    `json:"id" xorm:"pk autoincr"`
	Name string `json:"name"`
}

func (Tenant) GetAll(ctx context.Context) ([]Tenant, error) {
	var tenant []Tenant
	err := factory.DB(ctx).Find(&tenant)
	return tenant, err
}

type Namespace struct {
	Id         int    `json:"id" xorm:"pk autoincr"`
	TenantName string `json:"tenantName"`
	Name       string `json:"name"`
}

func (Namespace) GetByTenantName(ctx context.Context, tenantName string) ([]Namespace, error) {
	var namespace []Namespace
	err := factory.DB(ctx).Where("tenant_name = ?", tenantName).Find(&namespace)
	return namespace, err
}

func (Namespace) GetAll(ctx context.Context) ([]Namespace, error) {
	var namespace []Namespace
	err := factory.DB(ctx).Find(&namespace)
	return namespace, err
}

type DbAccount struct {
	TenantName string
	Name       string
	Host       string
	Port       int
	User       string
	Pwd        string
}

func (DbAccount) GetAll(ctx context.Context) ([]DbAccount, error) {
	var dbAccount []DbAccount
	err := factory.DB(ctx).Find(&dbAccount)
	return dbAccount, err
}

//registry.p2shop.com.cn/offer-api-pangpang-brand-qa
type ImageAccount struct {
	Registry  string `json:"registry"`
	LoginName string `json:"loginName"`
	Pwd       string `json:"pwd"`
}

func (ImageAccount) GetAll(ctx context.Context) ([]ImageAccount, error) {
	var imageAccount []ImageAccount
	err := factory.DB(ctx).Find(&imageAccount)
	return imageAccount, err
}
