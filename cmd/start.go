package cmd
import (
	"fmt"
)
func Start(){
	isContinue, serviceName, flag := (Flag{}).Init()
	if isContinue == false {
		return
	}
	fmt.Println(isContinue, serviceName, flag)
}