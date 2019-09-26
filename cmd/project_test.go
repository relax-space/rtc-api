package cmd_test

import (
	"rtc-api/cmd"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pangpanglabs/goutils/test"
)

func TestProject(t *testing.T) {
	// ps,err:=cmd.Project{}.GetAll()

	// fmt.Println(ps,err)

	cmd.SetEnv("")
	// names, err := cmd.Project{}.GetServiceNames("2")
	// fmt.Println(names, err)

	p, err := cmd.Project{}.GetProject("go-api")
	test.Ok(t, err)
	test.Assert(t, p != nil, "failure")
	cmd.ProjectOwner{}.ReLoad(p)
	spew.Dump(p, err)

}
