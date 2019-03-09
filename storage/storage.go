package storage

// Provider is an interface for retrieving images
type Provider interface {
	Get(id string) ([]byte, error)
}
