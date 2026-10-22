package repository

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gopkg.in/yaml.v3"
)

func pipelineYAML(orders ...int) string {
	out := "pipeline:\n"
	for _, o := range orders {
		out += "  - dataTypes: [syslog]\n"
		if o > 0 {
			out += "    order: " + strconv.Itoa(o) + "  # global pipeline order\n"
		}
		out += "    steps: []\n"
	}
	return out
}

func ordersOf(t *testing.T, content []byte) []int {
	t.Helper()
	var spec domain.PipelineSpec
	if err := yaml.Unmarshal(content, &spec); err != nil {
		t.Fatal(err)
	}
	out := make([]int, len(spec.Pipeline))
	for i, e := range spec.Pipeline {
		out[i] = int(e.Order)
	}
	return out
}

func TestPreserveDestOrder(t *testing.T) {
	cases := []struct {
		name     string
		existing []int
		incoming []int
		want     []int
	}{
		{"single entry keeps the operator's order", []int{5}, []int{10}, []int{5}},
		{"each entry keeps its own order", []int{7, 9}, []int{35, 36}, []int{7, 9}},
		{"orders copied onto every entry are replaced by the shipped ones", []int{35, 35}, []int{35, 36}, []int{35, 36}},
		{"a different number of entries takes the shipped file", []int{7}, []int{35, 36}, []int{35, 36}},
		{"an order of zero takes the shipped one", []int{0}, []int{12}, []int{12}},
	}

	for _, c := range cases {
		dest := filepath.Join(t.TempDir(), "f.yaml")
		if err := os.WriteFile(dest, []byte(pipelineYAML(c.existing...)), 0o644); err != nil {
			t.Fatal(err)
		}

		got := ordersOf(t, preserveDestOrder(dest, []byte(pipelineYAML(c.incoming...))))
		if len(got) != len(c.want) {
			t.Fatalf("%s: got %v, want %v", c.name, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: got %v, want %v", c.name, got, c.want)
				break
			}
		}
	}
}

func TestPreserveDestOrderWithoutAnExistingFile(t *testing.T) {
	incoming := []byte(pipelineYAML(35, 36))
	got := preserveDestOrder(filepath.Join(t.TempDir(), "missing.yaml"), incoming)
	if string(got) != string(incoming) {
		t.Errorf("content changed without an existing file:\n%s", got)
	}
}
