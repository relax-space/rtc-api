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
	dbAccounts, err := Project{}.GetAllDbAccount(jwtToken)
	if err != nil {
		return err
	}
	if p.Owner.IsMysql {
		folder += "/mysql"
		if err := (Folder{}).MkdirAll(folder); err != nil {
			return err
		}
		if err := d.writeMysql(dbAccounts, p.Owner.Databases[MYSQL.String()], folder); err != nil {
			return err
		}
	}
	Info("base data fetched")
	return nil
}
func (d BaseData) getDatabaseName(dbName, namespace string) string {
	dbNameNew := dbName
	if len(namespace) != 0 {
		dbNameNew = dbName + "_" + strings.Replace(namespace, "-", "_", -1)
	}
	return dbNameNew
}
func (d BaseData) writeMysql(dbAccounts []DbAccountDto, dbDtos []DatabaseDto, folder string) error {
	for _, dbDto := range dbDtos {
		dbAccount := Project{}.GetDbAccount(dbAccounts, MYSQL, dbDto.TenantName)
		config := mysql.NewConfig()
		config.User = dbAccount.User
		config.Passwd = dbAccount.Pwd
		config.DBName = d.getDatabaseName(dbDto.DbName, dbDto.Namespace)
		config.Net = "tcp"
		config.Addr = dbAccount.Host + ":" + fmt.Sprint(dbAccount.Port)
		db, err := sql.Open("mysql", config.FormatDSN())
		if err != nil {
			return err
		}
		dumpDir := folder
		dumpFilenameFormat := fmt.Sprintf("%s-20060102T150405", dbDto.DbName)
		dumper, err := mysqldump.Register(db, dbDto.DbName, dumpDir, dumpFilenameFormat)
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
