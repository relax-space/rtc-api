package cmd

import (
	"database/sql"
	"fmt"
	"path"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/relax-space/go-mysqldump"
)

type BaseData struct {
}

func (d BaseData) Write(p *Project, jwtToken string, integrationTest bool, dbNet *string) error {
	if StringPointCheck(dbNet) == false {
		*dbNet = TCPDBNET.String()
	}
	if *dbNet == LOCALDBNET.String() {
		return nil
	}

	if ContainString(EMPTYDBNET.List(), *dbNet) == false {
		return fmt.Errorf("error invalid,only support%v", EMPTYDBNET.List())
	}

	return d.write(p, jwtToken, integrationTest, *dbNet)

}

func (d BaseData) write(p *Project, jwtToken string, integrationTest bool, dbNet string) error {
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
		if err := d.writeMysql(dbAccounts, p.Owner.Databases[MYSQL.String()], folder, integrationTest, dbNet, jwtToken); err != nil {
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
func (d BaseData) writeMysql(dbAccounts []DbAccountDto, dbDtos []DatabaseDto, folder string, integrationTest bool, dbNet, jwtToken string) error {
	for _, dbDto := range dbDtos {
		dbAccount := Project{}.GetDbAccount(dbAccounts, MYSQL, dbDto.TenantName)
		if dbNet == TCPDBNET.String() {
			return d.tcp(dbAccount, dbDto, folder, integrationTest)
		} else if dbNet == HTTPDBNET.String() {
			return d.http(dbAccount, dbDto, folder, integrationTest, jwtToken)
		}
	}
	return nil
}

func (d BaseData) tcp(dbAccount DbAccountDto, dbDto DatabaseDto, folder string, integrationTest bool) error {

	config := mysql.NewConfig()
	config.DBName = d.getDatabaseName(dbDto.DbName, dbDto.Namespace)
	config.Net = "tcp"
	if integrationTest == true {
		config.User = dbAccount.TUser
		config.Passwd = dbAccount.TPwd
		config.Addr = dbAccount.THost + ":" + fmt.Sprint(dbAccount.TPort)
	} else {
		config.User = dbAccount.User
		config.Passwd = dbAccount.Pwd
		config.Addr = dbAccount.Host + ":" + fmt.Sprint(dbAccount.Port)
	}

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
	return nil
}

func (d BaseData) http(dbAccount DbAccountDto, dbDto DatabaseDto, folder string, integrationTest bool, jwtToken string) error {
	urlStr := fmt.Sprintf("%v/dbfiles?nsPrefix=%v&nsSuffix=%v&dbName=%v", env.RtcDbUrl, dbDto.TenantName, dbDto.Namespace, dbDto.DbName)
	dumpFilenameFormat := fmt.Sprintf("%s-20060102T150405.sql", dbDto.DbName)
	p := path.Join(folder, dumpFilenameFormat)
	fmt.Println(p)
	return File{}.WriteUrl(urlStr, p, jwtToken)
}
