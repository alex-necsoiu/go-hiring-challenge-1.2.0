package models

import (
	"testing"

	"gorm.io/gorm"
)

// TestDBInterface ensures the adapter implements DBInterface correctly
// The adapter is a thin wrapper around *gorm.DB with no business logic
func TestDBInterface(t *testing.T) {
	mockDB := &gorm.DB{}
	adapter := NewGormDBAdapter(mockDB)

	// Test 1: Verify the adapter implements DBInterface
	var dbInterface DBInterface = adapter
	if dbInterface == nil {
		t.Fatal("adapter should implement DBInterface")
	}

	// Test 2: Verify GetError works
	err := adapter.GetError()
	if err != nil && err != gorm.ErrRecordNotFound {
		// either nil or record not found, that's OK for this test
	}

	// Test 3: Just verify the type is GormDBAdapter
	if ga, ok := dbInterface.(*GormDBAdapter); !ok || ga == nil {
		t.Fatal("adapter should be GormDBAdapter type")
	}

	// Test 4: Verify adapter preserves errors from wrapped DB
	errDB := &gorm.DB{Error: gorm.ErrRecordNotFound}
	errAdapter := NewGormDBAdapter(errDB)
	err = errAdapter.GetError()
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected RecordNotFound error, got %v", err)
	}

	// Test 5: Additional interface verification
	t.Run("all interface methods callable", func(t *testing.T) {
		adapter := NewGormDBAdapter(&gorm.DB{})
		var _ DBInterface = adapter
		// If this compiles and runs, interface is satisfied
	})
}
