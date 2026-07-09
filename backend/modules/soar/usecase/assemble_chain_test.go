package usecase

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

func TestFlowCommandUnmarshalYAML_LegacyString(t *testing.T) {
	src := []byte("commands:\n  - net user \"$(x)\" /active:no\n  - {command: \"echo b\", condition: OnSuccess}\n")
	var wrap struct {
		Commands []FlowCommand `yaml:"commands"`
	}
	if err := yaml.Unmarshal(src, &wrap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(wrap.Commands) != 2 || wrap.Commands[0].Command != `net user "$(x)" /active:no` || wrap.Commands[0].Condition != nil {
		t.Fatalf("bare-string entry not decoded: %+v", wrap.Commands)
	}
	if wrap.Commands[1].Command != "echo b" || wrap.Commands[1].Condition == nil || *wrap.Commands[1].Condition != domain.ConditionOnSuccess {
		t.Fatalf("mapping entry not decoded: %+v", wrap.Commands[1])
	}
}

func TestAssembleChain(t *testing.T) {
	ok := domain.ConditionOnSuccess
	fail := domain.ConditionOnFailure
	always := domain.ConditionAlways

	cases := []struct {
		name string
		in   []FlowCommand
		want string
	}{
		{"empty", nil, ""},
		{"single command drops leading condition",
			[]FlowCommand{{Command: "a", Condition: &ok}},
			"a"},
		{"chain uses each entry's condition as the joiner from the previous",
			[]FlowCommand{{Command: "a"}, {Command: "b", Condition: &ok}, {Command: "c", Condition: &fail}, {Command: "d", Condition: &always}},
			"a && b || c ; d"},
		{"nil condition on non-first defaults to ;",
			[]FlowCommand{{Command: "a"}, {Command: "b"}},
			"a ; b"},
		{"empty commands are skipped without leaving stray operators",
			[]FlowCommand{{Command: "a"}, {Command: "", Condition: &ok}, {Command: "c", Condition: &fail}},
			"a || c"},
	}
	for _, tc := range cases {
		if got := assembleChain(tc.in); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}
