package plugin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"
)

// TypeRegistry maintains a mapping of type names to their concrete types
type TypeRegistry struct {
	mu    sync.RWMutex
	types map[string]reflect.Type
}

var (
	globalRegistry = &TypeRegistry{
		types: make(map[string]reflect.Type),
	}
)

// RegisterType registers a type with the global registry
func RegisterType(value interface{}) {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.types[t.String()] = t
}

// SerializeObject serializes an object to bytes
func SerializeObject(obj interface{}) ([]byte, string, error) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		return nil, "", fmt.Errorf("serialization error: %w", err)
	}

	return buf.Bytes(), t.String(), nil
}

// DeserializeObject reconstructs an object from bytes
func DeserializeObject(data []byte, typeName string) (interface{}, error) {
	globalRegistry.mu.RLock()
	t, exists := globalRegistry.types[typeName]
	globalRegistry.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("type %s not registered", typeName)
	}

	// Create a new instance of the type
	v := reflect.New(t).Interface()

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(v); err != nil {
		return nil, fmt.Errorf("deserialization error: %w", err)
	}

	return v, nil
}
