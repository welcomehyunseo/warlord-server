package server

type Metadata struct {
	*Data
}

func NewMetadata() *Metadata {
	return &Metadata{
		Data: NewData(),
	}
}
