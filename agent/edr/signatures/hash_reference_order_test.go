package signatures

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestReferenceHashPermutationNativeDuplicateSelection(t *testing.T) {
	data, err := os.ReadFile("testdata/hash_candidate_order.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Count  int `json:"count"`
			WeakAt int `json:"weak_at"`
			Winner int `json:"winner"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 45 {
		t.Fatal("wrong native control inventory")
	}
	digest := bytes.Repeat([]byte{71}, 16)
	for _, row := range fixture.Cases {
		order := make([]uint32, row.Count)
		for i := range order {
			order[i] = uint32(i)
		}
		if err := referenceHashPermutation(order, func(uint32) []byte { return digest }, int64(row.Count)*128); err != nil {
			t.Fatal(err)
		}
		// The pinned scanner searches an inclusive range and stops on its first
		// equal midpoint. These expected indices came from complete native scans.
		if got := int(order[(len(order)-1)/2]); got != row.Winner {
			t.Fatalf("count=%d weak_at=%d selected=%d native=%d", row.Count, row.WeakAt, got, row.Winner)
		}
	}
}

func TestReferenceHashPermutationBoundsAndBudget(t *testing.T) {
	digests := [][]byte{bytes.Repeat([]byte{1}, 16), bytes.Repeat([]byte{2}, 16), bytes.Repeat([]byte{3}, 16)}
	order := []uint32{0, 1, 2}
	if err := referenceHashPermutation(order, func(i uint32) []byte { return digests[i] }, 0); !errors.Is(err, errHashOrderBudget) {
		t.Fatalf("work limit: %v", err)
	}
	order = []uint32{0, 1, 2}
	if err := referenceHashPermutation(order, func(i uint32) []byte { return digests[i] }, 128); err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(order); i++ {
		if referenceHashCompare(digests[order[i-1]], digests[order[i]]) > 0 {
			t.Fatal("digest order")
		}
	}
	// The uniform optimization must remain linear on duplicate-heavy input.
	order = make([]uint32, 100000)
	for i := range order {
		order[i] = uint32(i)
	}
	if err := referenceHashPermutation(order, func(uint32) []byte { return digests[0] }, int64(len(order))); err != nil {
		t.Fatal(err)
	}
	if order[0] != 1 || order[len(order)-1] != 0 {
		t.Fatal("uniform native permutation")
	}
}

func TestReferenceHashIndexNativeNeighbours(t *testing.T) {
	data, err := os.ReadFile("testdata/hash_candidate_neighbours.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Algorithm string   `json:"algorithm"`
			Ordinal   int      `json:"ordinal"`
			Size      string   `json:"size"`
			Digests   []string `json:"digests"`
			Actual    string   `json:"actual"`
			Expected  int      `json:"expected_index"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 120 {
		t.Fatal("native neighbour inventory")
	}
	for _, row := range fixture.Cases {
		t.Run(fmt.Sprintf("%s/%d", row.Algorithm, row.Ordinal), func(t *testing.T) {
			size := int64(-1)
			if row.Size != "*" {
				var err error
				size, err = strconv.ParseInt(row.Size, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, split := range []bool{false, true} {
				h := newDetectionHashTable()
				for i, value := range row.Digests {
					digest, err := hex.DecodeString(value)
					if err != nil {
						t.Fatal(err)
					}
					target := h
					if split {
						target = newDetectionHashTable()
						h.shards = append(h.shards, target)
					}
					if !target.add(string(digest), size, fmt.Sprintf("Candidate.C%03d", i)) {
						t.Fatal("add")
					}
				}
				if err := h.freeze(); err != nil {
					t.Fatal(err)
				}
				actual, err := hex.DecodeString(row.Actual)
				if err != nil {
					t.Fatal(err)
				}
				lookupSize := size
				if lookupSize < 0 {
					lookupSize = 8192
				}
				if got, want := h.lookup(string(actual), lookupSize), fmt.Sprintf("Candidate.C%03d", row.Expected); got != want {
					t.Fatalf("split=%v got=%s native=%s", split, got, want)
				}
			}
		})
	}
}

func TestWholeHashSourcePriorityAndClassification(t *testing.T) {
	digest := bytes.Repeat([]byte{49}, 16)
	for _, row := range []struct{ first, second, want string }{
		{"Local.First", "Supplemental.Second", "Local.First"},
		{"Weak.First", "Strong.Second", "Strong.Second"},
		{"Strong.First", "Weak.Second", "Weak.Second"},
		{"PUA.First", "Strong.Second", "Strong.Second"},
	} {
		root := newWholeHashIndex()
		for _, name := range []string{row.first, row.second} {
			shard := newDetectionHashTable()
			shard.add(string(digest), 37, name)
			root.shards = append(root.shards, shard)
		}
		if err := root.freeze(); err != nil {
			t.Fatal(err)
		}
		if got := root.lookup(string(digest), 37); got != row.want {
			t.Errorf("%s/%s: got %s want %s", row.first, row.second, got, row.want)
		}
	}
}
