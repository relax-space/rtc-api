package cmd

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

	if err = (ProjectOwner{}).ReLoad(project); err != nil {
		Error(err)
		return
	}

	if err = (Folder{}).DeleteAll(TEMP_FILE); err != nil {
		Error(err)
		return
	}

	if err = (BaseData{}).Write(project); err != nil {
		Error(err)
		return
	}

	if err = (Nginx{}).Write(project); err != nil {
		Error(err)
		return
	}

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
