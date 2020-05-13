package pipeline

type IDGen interface {
	ID() (uint64, error)
}
