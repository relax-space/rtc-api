package cmd

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/relax-space/go-mysqldump"
)

type BaseData struct {
}

func (d BaseData) Write(p *Project) error {
	Info("base data fetching ...")
	folder := fmt.Sprintf("%v/database", TEMP_FILE)
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

func (d BaseData) writeMysql(dbAccount DbAccount, folder string) error {

	for _, dbName := range dbAccount.DbNames {
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
		dumper, err := mysqldump.Register(db, config.DBName, dumpDir, dumpFilenameFormat)
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
