package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type Xorm struct {
	Mysql     *xorm.Engine
	SqlServer *xorm.Engine
}

func (Xorm) InitSql(project *ProjectDto) (err error) {
	fmt.Println("sql data loading...")
	dbXorm := &Xorm{}
	dbXorm.Mysql, err = xorm.NewEngine("mysql", fmt.Sprintf("root:1234@tcp(127.0.0.1:%v)/mysql?charset=utf8", 3306))
	if err != nil {
		return
	}

	dbXorm.SqlServer, err = xorm.NewEngine("mssql",
		"server=127.0.0.1;user id=sa;password=Eland123;database=master;connection timeout=300")
	if err != nil {
		return
	}
	projects := []*ProjectDto{project}
	err = dbXorm.insertSql(projects)
	fmt.Println("sql data loaded.")
	return
}

func (d *Xorm) insertSql(projects []*ProjectDto) (err error) {
	for _, projectDto := range projects {
		if len(projectDto.ServiceName) == 0 {
			continue
		}
		if (ProjectInfo{}).ShouldDb(projectDto, MYSQL) {
			if err = d.insert(d.Mysql, projectDto, MYSQL); err != nil {
				return
			}
		}
		if (ProjectInfo{}).ShouldDb(projectDto, SQLSERVER) {
			if err = d.insert(d.SqlServer, projectDto, SQLSERVER); err != nil {
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

func (d *Xorm) insert(db *xorm.Engine, project *ProjectDto, dbType DateBaseType) (err error) {
	fileName := fmt.Sprintf("temp/%v/%v/*.sql", project.ServiceName, dbType.String())
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
