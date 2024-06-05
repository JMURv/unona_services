package memory

type Broker struct {
	msgs map[string][]byte
}

func New() *Broker {
	return &Broker{
		msgs: make(map[string][]byte),
	}
}
