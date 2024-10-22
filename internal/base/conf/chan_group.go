package conf

type ExitChanStruct struct {
	ExitChan chan int
}

func NewExitChanStruct() *ExitChanStruct {
	return &ExitChanStruct{ExitChan: make(chan int)}
}
