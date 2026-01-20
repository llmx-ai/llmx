package provider

import (
	"testing"
)

func TestRegister(t *testing.T) {
	// Register a test provider
	Register("test", func(opts map[string]interface{}) (Provider, error) {
		return nil, nil
	})

	if !IsRegistered("test") {
		t.Error("Expected test provider to be registered")
	}

	list := List()
	found := false
	for _, name := range list {
		if name == "test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected test provider in list")
	}
}

func TestNew_NotFound(t *testing.T) {
	_, err := New("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for nonexistent provider")
	}
}

func TestIsRegistered(t *testing.T) {
	// Register a provider
	Register("test2", func(opts map[string]interface{}) (Provider, error) {
		return nil, nil
	})

	if !IsRegistered("test2") {
		t.Error("Expected test2 provider to be registered")
	}

	if IsRegistered("nonexistent") {
		t.Error("Expected nonexistent provider to not be registered")
	}
}
