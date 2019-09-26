package cmd

import (
	_ "github.com/denisenkom/go-mssqldb"
)

func Start() {
	isContinue, serviceName, flag := (Flag{}).Init()
	if isContinue == false {
		return
	}

	project, err := Project{}.GetProject(*serviceName)
	if err != nil {
		Error(err)
		return
	}

	ProjectOwner{}.ReLoad(project)

	if err = (&Compose{}).Write(project, flag); err != nil {
		Error(err)
		return
	}

	if err = (&Compose{}).Exec(project, flag); err != nil {
		Error(err)
		return
	}

	Info("==> you can start testing now. check health by `docker ps -a`")
}
