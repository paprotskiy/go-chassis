package ports

type MsgBroker interface {
	PushToBroker(data []byte) error
}
