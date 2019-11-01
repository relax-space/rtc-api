package cmd

func Start(version string) {
	isContinue, serviceName, flag := (Flag{}).Init(version)
	if isContinue == false {
		return
	}

	if err := (Folder{}).DeleteLocalSql(TEMP_FILE, flag.LocalSql); err != nil {
		Error(err)
		return
	}

	// simple service
	if (ComposeSimple{}).ShouldSimple(*serviceName) {
		if err := (ComposeSimple{}).Start(*serviceName, flag); err != nil {
			Error(err)
			return
		}
		return
	}
	project, err := Project{}.GetProject(*serviceName, *flag.JwtToken)
	if err != nil {
		Error(err)
		return
	}

	if err = (BaseData{}).Write(project, *flag.JwtToken, *flag.LocalSql); err != nil {
		Error(err)
		return
	}

	if err = (Nginx{}).Write(project, *flag.Prefix); err != nil {
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
