package monobrew

import "fmt"

func PanicMsg(msg string) {
	msg = fmt.Sprintf("EXITING - %s\n", msg)
	panic(msg)
}

func PanicIfErr(e error) {
	if e != nil {
		panic(e)
	}
}

func WarnMsg(msg string) {
	fmt.Printf("WARNING - %s\n", msg)
}
