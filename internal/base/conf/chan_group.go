package conf

type ChanGroup struct {
	InternalCommandSend chan []byte
}

func NewChanGroup() *ChanGroup {
	return &ChanGroup{
		InternalCommandSend: make(chan []byte),
	}
}
