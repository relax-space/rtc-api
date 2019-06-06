package main

import (
	"bytes"
	"errors"
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
	defer dbXorm.Mysql.Close()

	projects := []*ProjectDto{project}
	if err = dbXorm.insertSql(projects, portDto); err != nil {
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

func (d *Xorm) insertSql(projects []*ProjectDto, portDto PortDto) (err error) {
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
			if len(projectDto.Databases) != 1 {
				err = errors.New("A microservice can only support one sqlserver database.")
				return
			}
			databaseName := projectDto.Databases[SQLSERVER.String()][0]
			if err = d.insertSqlserver(projectDto.ServiceName, portDto.SqlServer, databaseName); err != nil {
				return
			}
		}
		if len(projectDto.SubProjects) != 0 {
			if err = d.insertSql(projectDto.SubProjects, portDto); err != nil {
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
		if err = d.importView(db, f); err != nil {
			return
		}
	}
	return
}

func (d *Xorm) insertSqlserver(serviceName, port, databaseName string) (err error) {
	fileName := fmt.Sprintf("temp/%v/%v/*.sql", serviceName, SQLSERVER.String())
	files, err := filepath.Glob(fileName)
	if err != nil {
		return
	}
	db, err := d.initSqlserver(port, databaseName)
	if err != nil {
		return
	}
	defer db.Close()
	for _, f := range files {
		if err = d.importView(db, f); err != nil {
			return
		}
	}
	return
}
func (d *Xorm) initSqlserver(port, databaseName string) (db *xorm.Engine, err error) {
	db, err = xorm.NewEngine("mssql",
		fmt.Sprintf("driver={sql server};Server=127.0.0.1;port=%v;Database=%v;user id=sa;password=Eland123;Max Pool Size=2000;", port, databaseName))
	if err != nil {
		return
	}
	return
}

func (*Xorm) importView(db *xorm.Engine, fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	_, err = db.Import(bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	return nil
}
