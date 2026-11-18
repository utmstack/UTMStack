package netblock

import (
	"errors"
	"reflect"
	"testing"
)

func TestFilterSetRetainsFailedPartialAddAndRetriesCleanup(t *testing.T) {
	failed := errors.New("injected OS failure")
	removeFails := true
	var removed []uint64
	remove := func(id uint64) error {
		if removeFails {
			return failed
		}
		removed = append(removed, id)
		return nil
	}
	set, err := (filterSet{}).ensure(2, func(layer int) (uint64, error) {
		if layer == 1 {
			return 0, failed
		}
		return 41, nil
	}, remove)
	if err == nil || set.complete || !reflect.DeepEqual(set.ids, []uint64{41}) {
		t.Fatalf("partial filter lost or claimed complete: set=%+v err=%v", set, err)
	}
	adds := 0
	set, err = set.ensure(2, func(int) (uint64, error) { adds++; return 0, nil }, remove)
	if err == nil || adds != 0 || len(set.ids) != 1 {
		t.Fatal("retry added filters before failed cleanup completed")
	}
	removeFails = false
	set, err = set.ensure(2, func(layer int) (uint64, error) { return uint64(100 + layer), nil }, remove)
	if err != nil || !set.complete || !reflect.DeepEqual(set.ids, []uint64{100, 101}) || !reflect.DeepEqual(removed, []uint64{41}) {
		t.Fatalf("retry failed: set=%+v removed=%v err=%v", set, removed, err)
	}
	set, err = set.remove(func(id uint64) error {
		if id == 101 {
			return failed
		}
		return nil
	})
	if err == nil || set.complete || !reflect.DeepEqual(set.ids, []uint64{101}) {
		t.Fatalf("failed removal was discarded: set=%+v err=%v", set, err)
	}
	set, err = set.remove(func(uint64) error { return nil })
	if err != nil || len(set.ids) != 0 {
		t.Fatalf("cleanup retry: set=%+v err=%v", set, err)
	}
}
