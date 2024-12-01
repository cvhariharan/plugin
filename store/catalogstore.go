package store

import (
	"sync"
)

type SocketType string

const (
	TCP  SocketType = "tcp"
	UNIX SocketType = "unix"
)

type ServiceInfo struct {
	Address string
	Socket  SocketType
}

type CatalogStore interface {
	Add(name string, s ServiceInfo) bool
	Get(name string) (ServiceInfo, bool)
}

type MemCatalogStore struct {
	m  map[string]ServiceInfo
	mu sync.Mutex
}

func NewMemCatalogStore() CatalogStore {
	return &MemCatalogStore{
		m: make(map[string]ServiceInfo),
	}
}

func (m *MemCatalogStore) Add(name string, s ServiceInfo) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[name] = s
	return true
}

func (m *MemCatalogStore) Get(name string) (ServiceInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.m[name]
	return p, ok
}
