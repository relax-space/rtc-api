package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/relax-space/go-mysqldump"
)

type BaseData struct {
}

func (d BaseData) Write(p *Project, jwtToken string) error {
	Info("base data fetching ...")
	folder := fmt.Sprintf("%v/database", TEMP_FILE)
<<<<<<< HEAD
=======
	dbAccounts, err := Project{}.GetAllDbAccount(jwtToken)
	if err != nil {
		return err
	}
>>>>>>> 8b7cd30... #153 devloper self test
	if p.Owner.IsMysql {
		folder += "/mysql"
		if err := (Folder{}).MkdirAll(folder); err != nil {
			return err
		}
		if err := d.writeMysql(p.Owner.MysqlAccount, folder); err != nil {
			return err
		}
	}
	Info("base data fetched")
	return nil
}

func (d BaseData) getNamePure(name string) (string, error) {
	namespaces, err := d.namespaceFilters()
	if err != nil {
		return "", err
	}
	namePure := name
	for _, ns := range namespaces {
		if strings.HasSuffix(name, ns) {
			namePure = strings.TrimSuffix(name, ns)
			break
		}
	}
	return namePure, nil
}
<<<<<<< HEAD
func (d BaseData) namespaceFilters() ([]string, error) {
	namespaces, err := (Project{}).GetNamespace()
	if err != nil {
		return nil, err
	}
	nsNew := make([]string, 0)
	for _, ns := range namespaces {
		ns.Name = "_" + strings.Replace(ns.Name, "-", "_", -1)
		nsNew = append(nsNew, ns.Name)
	}
	return nsNew, nil
}

func (d BaseData) writeMysql(dbAccount DbAccountDto, folder string) error {

	for _, dbName := range dbAccount.DbNames {
		namePure, err := d.getNamePure(dbName)
		if err != nil {
			return err
		}

=======
func (d BaseData) writeMysql(dbAccounts []DbAccountDto, dbDtos []DatabaseDto, folder string) error {
	for _, dbDto := range dbDtos {
		dbAccount := Project{}.GetDbAccount(dbAccounts, MYSQL, dbDto.TenantName)
>>>>>>> 8b7cd30... #153 devloper self test
		config := mysql.NewConfig()
		config.User = dbAccount.User
		config.Passwd = dbAccount.Pwd
		config.DBName = dbName
		config.Net = "tcp"
		config.Addr = dbAccount.Host + ":" + fmt.Sprint(dbAccount.Port)
		db, err := sql.Open("mysql", config.FormatDSN())
		if err != nil {
			return err
		}
		dumpDir := folder
		dumpFilenameFormat := fmt.Sprintf("%s-20060102T150405", dbName)
		dumper, err := mysqldump.Register(db, namePure, dumpDir, dumpFilenameFormat)
		if err != nil {
			return err
		}
		err = dumper.Dump()
		if err != nil {
			return err
		}
		dumper.Close()
	}
	return nil
}
