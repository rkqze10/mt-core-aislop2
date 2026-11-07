package pipeline

import (
	"strings"
)

// PipelineStage represents a single command invocation in a pipeline.
type PipelineStage struct {
	CmdName string
	Args    []string
	IsPiped bool
}

// TokenizeLine splits a shell command line into tokens, respecting quotes and escapes.
func TokenizeLine(line string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escaped := false
	quoteEmitted := false

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if inQuotes {
			// In single quotes, backslash is literal and does not escape
			if quoteChar == '\'' {
				if r == '\'' {
					inQuotes = false
					quoteChar = 0
					quoteEmitted = true
				} else {
					current.WriteRune(r)
				}
				continue
			}

			// In double quotes, backslash escapes quoteChar and backslash
			if r == '\\' {
				if i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\') {
					escaped = true
					continue
				}
			}

			if r == quoteChar {
				inQuotes = false
				quoteChar = 0
				quoteEmitted = true
			} else {
				current.WriteRune(r)
			}
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		switch r {
		case '"', '\'':
			inQuotes = true
			quoteChar = r
		case '|':
			if current.Len() > 0 || quoteEmitted {
				tokens = append(tokens, current.String())
				current.Reset()
				quoteEmitted = false
			}
			tokens = append(tokens, "|")
		case ' ', '\t':
			if current.Len() > 0 || quoteEmitted {
				tokens = append(tokens, current.String())
				current.Reset()
				quoteEmitted = false
			}
		default:
			current.WriteRune(r)
		}
	}

	// Handle trailing dangling escape
	if escaped {
		current.WriteRune('\\')
	}

	if current.Len() > 0 || quoteEmitted {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// ParseExecutionLine converts a command line into structured pipeline stages.
func ParseExecutionLine(line string) []PipelineStage {
	tokens := TokenizeLine(strings.TrimSpace(line))
	if len(tokens) == 0 {
		return nil
	}

	var stages []PipelineStage
	var current *PipelineStage
	expectCmd := false

	for _, token := range tokens {
		if token == "|" {
			if current == nil {
				// Leading pipe or double pipe (e.g. "| cmd" or "cmd || cmd")
				// Mark an empty/invalid stage to prevent silent corruption
				stages = append(stages, PipelineStage{
					CmdName: "",
					Args:    nil,
					IsPiped: len(stages) > 0,
				})
			} else {
				stages = append(stages, *current)
				current = nil
			}
			expectCmd = true
			continue
		}

		expectCmd = false
		if current == nil {
			clean := strings.ToLower(strings.TrimPrefix(token, "/"))
			isPiped := len(stages) > 0
			current = &PipelineStage{
				CmdName: clean,
				Args:    []string{},
				IsPiped: isPiped,
			}
		} else {
			current.Args = append(current.Args, token)
		}
	}

	if current != nil {
		stages = append(stages, *current)
	} else if expectCmd {
		// Trailing pipe (e.g. "cmd |") - record empty trailing stage
		stages = append(stages, PipelineStage{
			CmdName: "",
			Args:    nil,
			IsPiped: true,
		})
	}

	return stages
}

