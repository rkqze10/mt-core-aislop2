package pipeline

import (
	"reflect"
	"testing"
)

func TestTokenizeLine(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    `/sh ls -la`,
			expected: []string{"/sh", "ls", "-la"},
		},
		{
			input:    `/sh echo "hello world"`,
			expected: []string{"/sh", "echo", "hello world"},
		},
		{
			input:    `/cmd | grep "error"`,
			expected: []string{"/cmd", "|", "grep", "error"},
		},
		{
			input:    `/sh echo ""`,
			expected: []string{"/sh", "echo", ""},
		},
		{
			input:    `/sh echo 'literal\backslash'`,
			expected: []string{"/sh", "echo", `literal\backslash`},
		},
		{
			input:    `/sh echo trailing\`,
			expected: []string{"/sh", "echo", `trailing\`},
		},
	}

	for _, tc := range tests {
		actual := TokenizeLine(tc.input)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("TokenizeLine(%q) = %v; want %v", tc.input, actual, tc.expected)
		}
	}
}

func TestParseExecutionLine(t *testing.T) {
	stages := ParseExecutionLine(`/sh cat log.txt | grep -i "error" | grep "fatal"`)
	if len(stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(stages))
	}

	if stages[0].CmdName != "sh" || len(stages[0].Args) != 2 || stages[0].IsPiped {
		t.Errorf("stage 0 invalid: %+v", stages[0])
	}

	if stages[1].CmdName != "grep" || len(stages[1].Args) != 2 || !stages[1].IsPiped {
		t.Errorf("stage 1 invalid: %+v", stages[1])
	}

	// Test pipe edge cases
	stagesEmpty := ParseExecutionLine("")
	if stagesEmpty != nil {
		t.Errorf("expected nil for empty line, got %+v", stagesEmpty)
	}

	stagesLeadingPipe := ParseExecutionLine("| grep foo")
	if len(stagesLeadingPipe) != 2 || stagesLeadingPipe[0].CmdName != "" {
		t.Errorf("expected empty stage for leading pipe, got %+v", stagesLeadingPipe)
	}

	stagesTrailingPipe := ParseExecutionLine("echo foo |")
	if len(stagesTrailingPipe) != 2 || stagesTrailingPipe[1].CmdName != "" {
		t.Errorf("expected empty trailing stage for trailing pipe, got %+v", stagesTrailingPipe)
	}
}
