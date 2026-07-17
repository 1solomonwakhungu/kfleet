package store

// Store defines the persistence operations used by the hub.
type Store interface {
	Close() error
}

type memoryStore struct{}

// Open returns a placeholder in-memory store. The database path is reserved
// for the persistent implementation.
func Open(dbPath string) (Store, error) {
	_ = dbPath
	return &memoryStore{}, nil
}

func (s *memoryStore) Close() error {
	return nil
}
