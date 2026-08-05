package repository

import (
	"fmt"
	"sync"
	"testing"
)

func TestSaveKeepsTheOtherTenantsAndTheDefault(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save("", &FileConfig{Provider: "default", Model: "d"}); err != nil {
		t.Fatalf("save default: %v", err)
	}
	if err := s.Save("tenant-a", &FileConfig{Provider: "a", Model: "ma"}); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := s.Save("tenant-b", &FileConfig{Provider: "b", Model: "mb"}); err != nil {
		t.Fatalf("save b: %v", err)
	}

	for _, tc := range []struct{ tenant, want string }{
		{"", "default"},
		{"tenant-a", "a"},
		{"tenant-b", "b"},
	} {
		got, err := s.Load(tc.tenant)
		if err != nil {
			t.Fatalf("load %q: %v", tc.tenant, err)
		}
		if got == nil || got.Provider != tc.want {
			t.Fatalf("load %q = %+v, want provider %q", tc.tenant, got, tc.want)
		}
	}
}

// A tenant with no configuration of its own runs on the instance default.
func TestLoadFallsBackToTheDefault(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save("", &FileConfig{Provider: "default"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := s.Load("tenant-with-nothing")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got == nil || got.Provider != "default" {
		t.Fatalf("load = %+v, want the default", got)
	}
}

// LoadOwn is how Update tells "inherits" from "has its own", so a tenant saving
// for the first time is not shown the default's masked secrets as if they were
// already theirs.
func TestLoadOwnDoesNotFallBack(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save("", &FileConfig{Provider: "default"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := s.LoadOwn("tenant-with-nothing")
	if err != nil {
		t.Fatalf("load own: %v", err)
	}
	if got != nil {
		t.Fatalf("load own = %+v, want nil", got)
	}
}

func TestLoadOnAnEmptyStore(t *testing.T) {
	got, err := NewConfigStore(t.TempDir()).Load("whatever")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != nil {
		t.Fatalf("load = %+v, want nil before anything is written", got)
	}
}

// The point of Delete: a tenant that saved once is otherwise forked forever, so
// a rotated key on the instance default would never reach it.
func TestDeleteGoesBackToInheriting(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save("", &FileConfig{Provider: "default"}); err != nil {
		t.Fatalf("save default: %v", err)
	}
	if err := s.Save("tenant-a", &FileConfig{Provider: "own"}); err != nil {
		t.Fatalf("save own: %v", err)
	}

	if got, _ := s.Load("tenant-a"); got == nil || got.Provider != "own" {
		t.Fatalf("before delete = %+v, want its own", got)
	}

	if err := s.Delete("tenant-a"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got, err := s.Load("tenant-a")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got == nil || got.Provider != "default" {
		t.Fatalf("after delete = %+v, want the instance default", got)
	}
	if own, _ := s.LoadOwn("tenant-a"); own != nil {
		t.Fatalf("still has its own: %+v", own)
	}

	// A later change to the default now reaches it, which is the whole point.
	if err := s.Save("", &FileConfig{Provider: "rotated"}); err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if got, _ := s.Load("tenant-a"); got == nil || got.Provider != "rotated" {
		t.Fatalf("after rotating the default = %+v, want it picked up", got)
	}
}

// Removing the instance default would leave every inheriting tenant with
// nothing, so it is refused rather than silently emptied.
func TestDeleteRefusesTheInstanceDefault(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save("", &FileConfig{Provider: "default"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := s.Delete(""); err == nil {
		t.Fatal("the instance default was removable")
	}
	if got, _ := s.Load(""); got == nil || got.Provider != "default" {
		t.Fatalf("default = %+v, want it untouched", got)
	}
}

func TestDeleteLeavesTheOtherTenantsAlone(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	_ = s.Save("", &FileConfig{Provider: "default"})
	_ = s.Save("tenant-a", &FileConfig{Provider: "a"})
	_ = s.Save("tenant-b", &FileConfig{Provider: "b"})

	if err := s.Delete("tenant-a"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if got, _ := s.LoadOwn("tenant-b"); got == nil || got.Provider != "b" {
		t.Fatalf("tenant-b = %+v, want it untouched", got)
	}
}

func TestDeleteOnATenantThatNeverSavedIsFine(t *testing.T) {
	s := NewConfigStore(t.TempDir())
	_ = s.Save("", &FileConfig{Provider: "default"})

	if err := s.Delete("never-saved"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

// Two replicas share the config directory through the same bind mount. Saving
// for a tenant is read-modify-write over a file holding every tenant, so
// without a lock that other processes see, the slower save erases the other
// tenant's entry.
func TestConcurrentSavesDoNotLoseATenant(t *testing.T) {
	dir := t.TempDir()

	// Separate stores: the in-process mutex is what a second replica does not
	// share, so one store per writer is what the deployment actually looks like.
	const writers = 8
	var wg sync.WaitGroup
	errs := make(chan error, writers)

	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := NewConfigStore(dir)
			errs <- s.Save(fmt.Sprintf("tenant-%d", i), &FileConfig{Model: fmt.Sprintf("m%d", i)})
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	s := NewConfigStore(dir)
	for i := range writers {
		got, err := s.LoadOwn(fmt.Sprintf("tenant-%d", i))
		if err != nil {
			t.Fatalf("LoadOwn: %v", err)
		}
		if got == nil {
			t.Errorf("tenant-%d was lost", i)
			continue
		}
		if want := fmt.Sprintf("m%d", i); got.Model != want {
			t.Errorf("tenant-%d model = %q, want %q", i, got.Model, want)
		}
	}
}
