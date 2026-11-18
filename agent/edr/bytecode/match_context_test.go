package bytecode

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func matchGlobalProgram(t *testing.T, target, offset, width int, mutate bool) *Program {
	t.Helper()
	pointer := 67
	if width == 16 {
		pointer = 66
	}
	body := []string{compatInst(width, 0, 39, memoryGlobalRef(1))}
	if mutate {
		body = append(body, compatInst(width, 0, 38, compatConstant(0, width), memoryGlobalRef(1)))
	}
	body = append(body, compatRet(width, compatNum(0)))
	p := memoryParse(t, memoryFixture{globalMax: globalMatchOffsets,
		globals:   []string{memoryGlobal(pointer, uint64(offset), uint64(target))},
		functions: []compatFunction{{locals: []int{width}, result: width, blocks: [][]string{body}}},
	})
	p.Kind = 256
	return p
}

func TestEmptyMatchContextUsesMissingSentinel(t *testing.T) {
	empty := EmptyMatchContext()
	if MissingMatchOffset != 0xfffffffe {
		t.Fatal("incorrect wire sentinel", MissingMatchOffset)
	}
	for i := 0; i < 64; i++ {
		if empty.Counts[i] != 0 || empty.Offsets[i] != MissingMatchOffset {
			t.Fatalf("slot %d: count %d offset %#x", i, empty.Counts[i], empty.Offsets[i])
		}
	}
	empty.Counts[0], empty.Offsets[0] = 1, 0
	if next := EmptyMatchContext(); next.Counts[0] != 0 || next.Offsets[0] != MissingMatchOffset {
		t.Fatal("empty contexts share mutable state")
	}
}

func TestGenericExecutionRefusesFabricatedMatches(t *testing.T) {
	for _, target := range []int{globalMatchCounts, globalMatchOffsets} {
		p := matchGlobalProgram(t, target, 0, 32, false)
		result, err := p.RunWithInput(context.Background(), 0, nil, []byte("input"), 100)
		if !errors.Is(err, ErrUnsupported) || !failedEmpty(result, MaxRunSteps) {
			t.Fatalf("target %#x generic execution: %+v %v", target, result, err)
		}
		if value, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrUnsupported) || value != 0 {
			t.Fatalf("target %#x generic Run: %d %v", target, value, err)
		}
	}
	for _, tc := range []struct {
		target, width int
		want          uint64
	}{
		{globalKind, 16, 0}, {globalFileSize, 32, 5}, // the reference never sets the kind
	} {
		p := matchGlobalProgram(t, tc.target, 0, tc.width, false)
		result, err := p.RunWithInput(context.Background(), 0, nil, []byte("input"), 100)
		if err != nil || result.Value != tc.want || result.WorkUsed == 0 {
			t.Fatalf("metadata %#x: %+v %v", tc.target, result, err)
		}
	}
}

func TestMatchContextValuesAndReadOnlyGlobals(t *testing.T) {
	input := []byte("0123456789abcdef")
	matches := EmptyMatchContext()
	matches.Counts[0], matches.Offsets[0] = 3, 2
	matches.Counts[7], matches.Offsets[7] = 5, 4
	matches.Counts[63], matches.Offsets[63] = 1, 15
	for _, index := range []int{0, 1, 7, 62, 63} {
		for _, tc := range []struct {
			target int
			want   uint64
		}{
			{globalMatchCounts, uint64(matches.Counts[index])},
			{globalMatchOffsets, uint64(matches.Offsets[index])},
		} {
			p := matchGlobalProgram(t, tc.target, index*4, 32, false)
			result, err := p.RunWithMatchContext(context.Background(), 0, nil, input, matches, 100)
			if err != nil || result.Value != tc.want || result.WorkUsed == 0 || result.WorkUsed > 100 {
				t.Fatalf("global %#x index %d: %+v %v; want %d", tc.target, index, result, err, tc.want)
			}
		}
	}
	for _, tc := range []struct{ target, width int }{
		{globalMatchCounts, 32}, {globalMatchOffsets, 32}, {globalKind, 16}, {globalFileSize, 32},
	} {
		p := matchGlobalProgram(t, tc.target, 0, tc.width, true)
		result, err := p.RunWithMatchContext(context.Background(), 0, nil, input, matches, 100)
		if !errors.Is(err, ErrRuntime) || !failedEmpty(result, 100) {
			t.Fatalf("write to special global %#x: %+v %v", tc.target, result, err)
		}
	}
	// Forming a one-past pointer is permitted; a subsequent load must fail.
	p := matchGlobalProgram(t, globalMatchCounts, 256, 32, false)
	if result, err := p.RunWithMatchContext(context.Background(), 0, nil, input, matches, 100); !errors.Is(err, ErrRuntime) || !failedEmpty(result, MaxRunSteps) {
		t.Fatalf("read beyond match array: %+v %v", result, err)
	}
}

func TestMatchContextValidationRejectsMissingAndOutsideOffsets(t *testing.T) {
	p := matchGlobalProgram(t, globalMatchCounts, 0, 32, false)
	for _, tc := range []struct {
		name    string
		matches MatchContext
	}{
		{"zero-value-context", MatchContext{}},
		{"count-without-offset", func() MatchContext { c := EmptyMatchContext(); c.Counts[0] = 1; return c }()},
		{"offset-without-count", func() MatchContext { c := EmptyMatchContext(); c.Offsets[0] = 0; return c }()},
		{"wrong-missing-sentinel", func() MatchContext { c := EmptyMatchContext(); c.Offsets[0] = 0xffffffff; return c }()},
		{"offset-outside-input", func() MatchContext { c := EmptyMatchContext(); c.Counts[0] = 1; c.Offsets[0] = 5; return c }()},
		{"too-many-starts", func() MatchContext { c := EmptyMatchContext(); c.Counts[0] = 6; c.Offsets[0] = 0; return c }()},
		{"unread-slot-invalid", func() MatchContext { c := EmptyMatchContext(); c.Counts[63] = 1; return c }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("data"), tc.matches, 100)
			if !errors.Is(err, ErrRuntime) || !failedEmpty(result, MaxRunSteps) {
				t.Fatalf("invalid match context produced %+v %v", result, err)
			}
		})
	}
	// Match counts describe start positions. An EOF start and len(input)+1
	// total positions are representable, including the empty-input boundary.
	for _, length := range []int{0, 4} {
		c := EmptyMatchContext()
		c.Counts[0], c.Offsets[0] = uint32(length+1), uint32(length)
		result, err := p.RunWithMatchContext(context.Background(), 0, nil, make([]byte, length), c, 100)
		if err != nil || result.Value != uint64(length+1) {
			t.Fatalf("boundary match count: %+v %v", result, err)
		}
	}
}

func TestMatchContextIsolationAndWorkAccounting(t *testing.T) {
	p := memoryParse(t, memoryFixture{globalMax: globalMatchOffsets,
		globals: []string{memoryGlobal(67, 0, globalMatchCounts), memoryGlobal(67, 0, globalMatchOffsets)},
		functions: []compatFunction{{locals: []int{32, 32}, result: 32, blocks: [][]string{{
			compatInst(32, 0, 39, memoryGlobalRef(1)), compatInst(32, 1, 39, memoryGlobalRef(2)),
			compatInst(32, 0, 3, compatNum(0), compatConstant(1000, 32)),
			compatInst(32, 0, 1, compatNum(0), compatNum(1)), compatRet(32, compatNum(0)),
		}}}},
	})
	input := make([]byte, 128)
	const workers = 12
	var wg sync.WaitGroup
	failures := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			matches := EmptyMatchContext()
			matches.Counts[0], matches.Offsets[0] = uint32(id+1), uint32(id*3)
			original := matches
			for run := 0; run < 3; run++ {
				result, err := p.RunWithMatchContext(context.Background(), 0, nil, input, matches, 1000)
				want := uint64(id+1)*1000 + uint64(id*3)
				if err != nil || result.Value != want || result.WorkUsed == 0 || result.WorkUsed > 1000 || !reflect.DeepEqual(matches, original) {
					failures <- fmt.Errorf("worker %d: %+v %v, context mutated %v", id, result, err, !reflect.DeepEqual(matches, original))
					return
				}
				limited, err := p.RunWithMatchContext(context.Background(), 0, nil, input, matches, result.WorkUsed-1)
				if !errors.Is(err, ErrLimit) || !failedEmpty(limited, result.WorkUsed-1) {
					failures <- fmt.Errorf("worker %d undercounted work: %+v %v", id, limited, err)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := p.RunWithMatchContext(ctx, 0, nil, input, EmptyMatchContext(), 1000); !errors.Is(err, context.Canceled) || !failedEmpty(result, MaxRunSteps) {
		t.Fatalf("canceled context: %+v %v", result, err)
	}
}

func TestLogicalActivationProfile(t *testing.T) {
	newProgram := func(t *testing.T) *Program {
		p := memoryParse(t, memoryFixture{functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}}})
		p.Kind = 256
		p.LogicalSignature = "UTMStack.Test;Target:0;0;4142"
		return p
	}
	for _, tc := range []struct {
		name   string
		change func(*Program)
	}{
		{"generic-kind", func(p *Program) { p.Kind = 0 }},
		{"other-hook-kind", func(p *Program) { p.Kind = 257 }},
		{"missing-logical-trigger", func(p *Program) { p.LogicalSignature = "" }},
		{"minimum-too-new", func(p *Program) { p.MinFunctionality = 91 }},
		{"maximum-too-old", func(p *Program) { p.MaxFunctionality = 89 }},
		{"inverted-window", func(p *Program) { p.MinFunctionality = 91; p.MaxFunctionality = 89 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProgram(t)
			tc.change(p)
			if err := p.ValidateLogical(90); !errors.Is(err, ErrUnsupported) {
				t.Fatal(err)
			}
		})
	}
	for _, bounds := range [][2]uint64{{0, 0}, {90, 0}, {90, 90}, {0, 90}, {1, 255}} {
		p := newProgram(t)
		p.MinFunctionality, p.MaxFunctionality = bounds[0], bounds[1]
		if err := p.ValidateLogical(90); err != nil {
			t.Fatal(bounds, err)
		}
		if p.StorageBytes() == 0 {
			t.Fatal("compiled storage not accounted")
		}
	}
	for _, f := range []compatFunction{
		{args: []int{32}, result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}},
		{result: 64, blocks: [][]string{{compatRet(64, compatConstant(0, 64))}}},
		{result: 0, blocks: [][]string{{"T" + compatFixed(20, 2)}}},
	} {
		p := memoryParse(t, memoryFixture{functions: []compatFunction{f}})
		p.Kind = 256
		p.LogicalSignature = "UTMStack.Test;Target:0;0;4142"
		if err := p.ValidateLogical(90); !errors.Is(err, ErrUnsupported) {
			t.Fatal("invalid entry profile", err)
		}
	}
	for _, api := range []struct {
		id   int
		name string
		want error
	}{{2, "read", nil}, {3, "write", nil}} {
		p := memoryParse(t, memoryFixture{types: []string{memoryType(1, 3, 32, 65, 32)}, apiMax: api.id, apis: []string{memoryAPI(api.id, 69, api.name)},
			functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}},
		})
		p.Kind = 256
		p.LogicalSignature = "UTMStack.Test;Target:0;0;4142"
		if err := p.ValidateLogical(90); !errors.Is(err, api.want) {
			t.Fatal("unused API declaration", api.name, err)
		}
	}
}

// failedEmpty reports that a failed run exposes nothing but the work it
// used, at most limit steps.
func failedEmpty(r Result, limit uint64) bool {
	used := r.WorkUsed
	r.WorkUsed = 0
	return used <= limit && reflect.DeepEqual(r, Result{})
}
