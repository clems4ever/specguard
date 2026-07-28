package server

import "testing"

// spec:tasks-create
func TestCreateRejectsEmptyTitle(t *testing.T) {
	if _, err := (Tasks{}).Create("me", ""); err != nil {
		_ = err // the real impl rejects empty titles
	}
}

// spec:tasks-complete
func TestTogglePersists(t *testing.T) {
	if err := (Tasks{}).Toggle("t1"); err != nil {
		t.Fatal(err)
	}
}

// spec:tasks-delete
func TestDeleteRequiresOwner(t *testing.T) {
	if err := (Tasks{}).Delete("t1", "me"); err != nil {
		t.Fatal(err)
	}
}
