package models

import (
	"time"

	"github.com/go-xorm/xorm"
)

func InitDB(driver, connection string) (*xorm.Engine, error) {
	db, err := xorm.NewEngine(driver, connection)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(10 * time.Second) //How long will the idle connection in the connection pool remain? 连接池中的空闲连接保持多久
	db.SetMaxOpenConns(30)                  //Maximum number of connections allowed to open 允许打开的最大连接数
	db.SetMaxIdleConns(10)                  //Several idle connections remain in the connection pool when all connections are idle 所有连接空闲时，连接池中保留几个空闲连接
	db.Sync2(new(Relation))
	return db, nil
}

func InitTable(db *xorm.Engine) error {
	if err := db.Sync(new(Relation)); err != nil {
		return err
	}
	return nil
}

func DropTables(db *xorm.Engine) error {
	return db.DropTables(new(Relation))
}
