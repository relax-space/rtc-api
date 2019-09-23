package cmd_test

import(
	"rtc-api/cmd"
	"fmt"
	"testing"
)

func TestProject(t *testing.T){
	// ps,err:=cmd.Project{}.GetAll()

	// fmt.Println(ps,err)

	// names,err :=cmd.Project{}.GetServiceNames("2")
	// fmt.Println(names,err)

	p,err := cmd.Project{}.GetProject("go-api|")
	fmt.Println(p,err)
}