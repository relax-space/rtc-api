package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/go-xorm/xorm"
)

type Xorm struct {
	Mysql     *xorm.Engine
	SqlServer *xorm.Engine
}

func (Xorm) InitSql(project *ProjectDto, portDto PortDto) (err error) {
	log.Println("sql data loading...")
	dbXorm := &Xorm{}
	dbXorm.Mysql, err = xorm.NewEngine("mysql", fmt.Sprintf("root:1234@tcp(127.0.0.1:%v)/mysql?charset=utf8", portDto.Mysql))
	if err != nil {
		return
	}
	dbXorm.SqlServer, err = xorm.NewEngine("mssql",
		fmt.Sprintf("driver={sql server};Server=127.0.0.1;port=%v;Database=master;user id=sa;password=Eland123;Max Pool Size=2000;", portDto.SqlServer))
	if err != nil {
		return
	}

	projects := []*ProjectDto{project}
	if err = dbXorm.insertSql(projects); err != nil {
		return
	}
	if (ProjectInfo{}).ShouldEventBroker(project) {
		if err = dbXorm.insertSqlEventBroker(); err != nil {
			return
		}
	}
	log.Println("sql data loaded.")
	return
}

func (d *Xorm) insertSql(projects []*ProjectDto) (err error) {
	for _, projectDto := range projects {
		if len(projectDto.ServiceName) == 0 {
			continue
		}
		if (Database{}).ShouldDb(projectDto, MYSQL) {
			if err = d.insert(d.Mysql, projectDto.ServiceName, MYSQL); err != nil {
				return
			}
		}
		if (Database{}).ShouldDb(projectDto, SQLSERVER) {
			if err = d.insert(d.SqlServer, projectDto.ServiceName, SQLSERVER); err != nil {
				return
			}
		}
		if len(projectDto.SubProjects) != 0 {
			if err = d.insertSql(projectDto.SubProjects); err != nil {
				return
			}
		}
	}
	return
}

func (d *Xorm) insertSqlEventBroker() (err error) {
	return d.insert(d.Mysql, "kafka-consumer", MYSQL)
}

func (d *Xorm) insert(db *xorm.Engine, serviceName string, dbType DateBaseType) (err error) {
	fileName := fmt.Sprintf("temp/%v/%v/*.sql", serviceName, dbType.String())
	files, err := filepath.Glob(fileName)
	if err != nil {
		return
	}
	for _, f := range files {
		if err = d.importView(db, f, dbType); err != nil {
			return
		}
	}
	return
}

func (*Xorm) importView(db *xorm.Engine, fileName string, dbType DateBaseType) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	if dbType == SQLSERVER {
		_, err = db.Exec(string(b))
		if err != nil {
			return err
		}
	} else {
		_, err = db.Import(bytes.NewBuffer(b))
		if err != nil {
			return err
		}
	}

	return nil
}
