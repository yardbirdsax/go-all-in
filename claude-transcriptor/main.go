// claude-transcriptor converts a Claude Code session JSONL file to a Markdown interview transcript.
//
// Usage: claude-transcriptor <session.jsonl> [output.md]
//
// Session files are found in:
//
//	~/.claude/projects/<project-slug>/<session-id>.jsonl
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const usageText = `claude_transcriptor - Convert a Claude Code session JSONL file to a Markdown interview transcript.

Usage: claude_transcriptor <session.jsonl> [output.md]

Session files are found in:
  ~/.claude/projects/<project-slug>/<session-id>.jsonl
`

type exchange struct {
	speaker   string
	message   string
	timestamp string
}

// ContentItem represents one element of a message's content array.
type ContentItem struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error"`
	Content   string          `json:"content"`
}

// Message has a content field that is either a JSON string or a JSON array.
type Message struct {
	ContentStr   string
	ContentItems []ContentItem
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var raw struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Content) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw.Content, &s); err == nil {
		m.ContentStr = s
		return nil
	}
	var items []ContentItem
	if err := json.Unmarshal(raw.Content, &items); err == nil {
		m.ContentItems = items
	}
	return nil
}

type progressMessage struct {
	Type    string   `json:"type"`
	Message *Message `json:"message"`
}

type progressData struct {
	Message progressMessage `json:"message"`
}

type entry struct {
	Type          string          `json:"type"`
	Timestamp     string          `json:"timestamp"`
	Message       *Message        `json:"message"`
	ToolUseResult json.RawMessage `json:"toolUseResult"`
	Data          *progressData   `json:"data"`
}

// orderedStringValues extracts string values from a JSON object in insertion order.
func orderedStringValues(data json.RawMessage) []string {
	if len(data) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return nil
	}
	var values []string
	for dec.More() {
		if _, err := dec.Token(); err != nil { // discard key
			break
		}
		var v string
		if err := dec.Decode(&v); err != nil {
			break
		}
		values = append(values, v)
	}
	return values
}

type askQuestionInput struct {
	Questions []struct {
		Question string `json:"question"`
		Options  []struct {
			Label       string `json:"label"`
			Description string `json:"description"`
		} `json:"options"`
	} `json:"questions"`
}

func formatAskQuestion(input json.RawMessage) []string {
	var qi askQuestionInput
	if err := json.Unmarshal(input, &qi); err != nil {
		return nil
	}
	var parts []string
	for _, q := range qi.Questions {
		formatted := fmt.Sprintf("**%s**\n\n", q.Question)
		for _, opt := range q.Options {
			formatted += fmt.Sprintf("- **%s**: %s\n", opt.Label, opt.Description)
		}
		parts = append(parts, formatted)
	}
	return parts
}

func parseSession(sessionFile string) ([]exchange, error) {
	f, err := os.Open(sessionFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []*entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var e entry
		if err := json.Unmarshal(line, &e); err == nil {
			entries = append(entries, &e)
		} else {
			entries = append(entries, nil)
		}
	}

	// Pre-pass: collect ExitPlanMode tool use IDs.
	exitPlanIDs := make(map[string]bool)
	for _, e := range entries {
		if e == nil || e.Type != "assistant" || e.Message == nil {
			continue
		}
		for _, item := range e.Message.ContentItems {
			if item.Name == "ExitPlanMode" && item.ID != "" {
				exitPlanIDs[item.ID] = true
			}
		}
	}

	var transcript []exchange

	for _, e := range entries {
		if e == nil {
			continue
		}
		ts := e.Timestamp

		switch e.Type {

		case "user":
			if e.Message != nil {
				// String content.
				if c := e.Message.ContentStr; c != "" && !strings.HasPrefix(c, "<") {
					c = strings.ReplaceAll(c, "\\\n", "\n")
					transcript = append(transcript, exchange{"Josh", c, ts})
				}

				// List content — look for ExitPlanMode tool results.
				for _, item := range e.Message.ContentItems {
					if item.Type != "tool_result" || !exitPlanIDs[item.ToolUseID] {
						continue
					}
					raw := item.Content
					if item.IsError {
						feedback := ""
						if strings.Contains(raw, "the user said:\n") {
							parts := strings.SplitN(raw, "the user said:\n", 2)
							feedback = strings.TrimSpace(parts[1])
						}
						if feedback != "" {
							transcript = append(transcript, exchange{"Josh", fmt.Sprintf("**[Plan rejected]**\n\n%s", feedback), ts})
						} else if strings.Contains(raw, "STOP") {
							transcript = append(transcript, exchange{"Josh", "**[Plan accepted — context cleared]**", ts})
						} else {
							transcript = append(transcript, exchange{"Josh", "**[Plan rejected]**", ts})
						}
					} else {
						transcript = append(transcript, exchange{"Josh", "**[Plan accepted]**", ts})
					}
				}

				// toolUseResult handling.
				if raw := e.ToolUseResult; len(raw) > 0 && string(raw) != "null" {
					trimmed := bytes.TrimSpace(raw)
					if len(trimmed) > 0 && trimmed[0] == '{' {
						// It's a JSON object — look for answers.
						var resultObj struct {
							Answers json.RawMessage `json:"answers"`
						}
						if err := json.Unmarshal(raw, &resultObj); err == nil {
							if values := orderedStringValues(resultObj.Answers); len(values) > 0 {
								combined := strings.Join(values, "\n\n")
								transcript = append(transcript, exchange{"Josh", combined, ts})
							}
						}
					} else {
						// Try as string.
						var resultStr string
						if err := json.Unmarshal(raw, &resultStr); err == nil {
							if strings.HasPrefix(resultStr, "Error:") {
								for _, item := range e.Message.ContentItems {
									if item.Type == "tool_result" && item.IsError {
										if strings.Contains(resultStr, "wants to clarify") {
											transcript = append(transcript, exchange{"Josh", "<requested clarification>", ts})
										}
										break
									}
								}
							}
						}
					}
				}
			}

		case "assistant":
			if e.Message == nil {
				continue
			}
			for _, item := range e.Message.ContentItems {
				switch item.Type {
				case "text":
					if strings.TrimSpace(item.Text) != "" {
						text := strings.ReplaceAll(item.Text, "\\\n", "\n")
						transcript = append(transcript, exchange{"Claude", text, ts})
					}
				case "tool_use":
					switch item.Name {
					case "AskUserQuestion":
						for _, formatted := range formatAskQuestion(item.Input) {
							transcript = append(transcript, exchange{"Claude", formatted, ts})
						}
					case "ExitPlanMode":
						var input struct {
							Plan string `json:"plan"`
						}
						if err := json.Unmarshal(item.Input, &input); err == nil && strings.TrimSpace(input.Plan) != "" {
							msg := fmt.Sprintf("**[Proposed Plan]**\n\n<details>\n<summary>View plan</summary>\n\n%s\n\n</details>", input.Plan)
							transcript = append(transcript, exchange{"Claude", msg, ts})
						}
					}
				}
			}

		case "progress":
			if e.Data == nil {
				continue
			}
			dataMsg := e.Data.Message
			if dataMsg.Message == nil {
				continue
			}
			for _, item := range dataMsg.Message.ContentItems {
				if item.Type != "text" {
					continue
				}
				text := item.Text
				if strings.TrimSpace(text) == "" {
					continue
				}
				switch dataMsg.Type {
				case "user":
					if !strings.HasPrefix(text, "<") {
						text = strings.ReplaceAll(text, "\\\n", "\n")
						transcript = append(transcript, exchange{"Claude", fmt.Sprintf("**[To subagent]** %s", text), ts})
					}
				case "assistant":
					text = strings.ReplaceAll(text, "\\\n", "\n")
					transcript = append(transcript, exchange{"Claude", text, ts})
				}
			}
		}
	}

	sort.SliceStable(transcript, func(i, j int) bool {
		return transcript[i].timestamp < transcript[j].timestamp
	})

	return transcript, nil
}

func writeTranscript(transcript []exchange, outputFile string) error {
	output := "# Interview Transcript: Claude Code Session\n\n"
	for _, e := range transcript {
		output += fmt.Sprintf("**%s:** %s\n\n", e.speaker, e.message)
	}
	return os.WriteFile(outputFile, []byte(output), 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usageText)
		os.Exit(1)
	}

	sessionFile := os.Args[1]
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", sessionFile)
		os.Exit(1)
	}

	outputFile := "transcript.md"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	transcript, err := parseSession(sessionFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := writeTranscript(transcript, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transcript written with %d exchanges to %s\n", len(transcript), outputFile)
}
