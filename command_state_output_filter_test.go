package main

import "testing"

func TestCommandStateOutputFilterConsumesRawOscBel(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("before\x1b]777;tui-helper;command-start;bnBtIHRlc3Q=\aafter")

	if result.Data != "beforeafter" {
		t.Fatalf("Data = %q, want beforeafter", result.Data)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].Type != "command-start" || result.Events[0].Command != "npm test" {
		t.Fatalf("Event = %#v, want command-start npm test", result.Events[0])
	}
}

func TestCommandStateOutputFilterConsumesTodoAIRawOscBel(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("before\x1b]777;todoai;command-start;bnBtIHRlc3Q=\aafter")

	if result.Data != "beforeafter" {
		t.Fatalf("Data = %q, want beforeafter", result.Data)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].Type != "command-start" || result.Events[0].Command != "npm test" {
		t.Fatalf("Event = %#v, want command-start npm test", result.Events[0])
	}
}

func TestCommandStateOutputFilterConsumesRawOscSt(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("before\x1b]777;tui-helper;command-end\x1b\\after")

	if result.Data != "beforeafter" {
		t.Fatalf("Data = %q, want beforeafter", result.Data)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].Type != "command-end" {
		t.Fatalf("Event = %#v, want command-end", result.Events[0])
	}
}

func TestCommandStateOutputFilterConsumesWindowsTextFallback(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("prefix 777;tui-helper;command-start;Y29kZXg=\r\nsuffix")

	if result.Data != "prefix suffix" {
		t.Fatalf("Data = %q, want prefix suffix", result.Data)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].Type != "command-start" || result.Events[0].Command != "codex" {
		t.Fatalf("Event = %#v, want command-start codex", result.Events[0])
	}
}

func TestCommandStateOutputFilterConsumesTodoAIWindowsTextFallback(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("prefix 777;todoai;command-start;Y29kZXg=\r\nsuffix")

	if result.Data != "prefix suffix" {
		t.Fatalf("Data = %q, want prefix suffix", result.Data)
	}
	if len(result.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(result.Events))
	}
	if result.Events[0].Type != "command-start" || result.Events[0].Command != "codex" {
		t.Fatalf("Event = %#v, want command-start codex", result.Events[0])
	}
}

func TestCommandStateOutputFilterPreservesBracketedWindowsTextFallback(t *testing.T) {
	filter := newCommandStateOutputFilter()
	input := "prefix ]777;tui-helper;command-start;Y2FsYw==\asuffix"

	result := filter.Filter(input)

	if result.Data != input {
		t.Fatalf("Data = %q, want %q", result.Data, input)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterConsumesSplitPayload(t *testing.T) {
	filter := newCommandStateOutputFilter()

	first := filter.Filter("before\x1b]777;todoai;command-start;")
	second := filter.Filter("Y2xhdWRl\aafter")

	if first.Data != "before" {
		t.Fatalf("first Data = %q, want before", first.Data)
	}
	if len(first.Events) != 0 {
		t.Fatalf("first Events = %#v, want none", first.Events)
	}
	if second.Data != "after" {
		t.Fatalf("second Data = %q, want after", second.Data)
	}
	if len(second.Events) != 1 {
		t.Fatalf("second Events length = %d, want 1", len(second.Events))
	}
	if second.Events[0].Type != "command-start" || second.Events[0].Command != "claude" {
		t.Fatalf("Event = %#v, want command-start claude", second.Events[0])
	}
}

func TestCommandStateOutputFilterConsumesSplitWindowsTextFallbackPrefix(t *testing.T) {
	filter := newCommandStateOutputFilter()

	first := filter.Filter("before 777;todo")
	second := filter.Filter("ai;command-start;Y29kZXg=\r\nafter")

	if first.Data != "before " {
		t.Fatalf("first Data = %q, want before ", first.Data)
	}
	if len(first.Events) != 0 {
		t.Fatalf("first Events = %#v, want none", first.Events)
	}
	if second.Data != "after" {
		t.Fatalf("second Data = %q, want after", second.Data)
	}
	if len(second.Events) != 1 {
		t.Fatalf("second Events length = %d, want 1", len(second.Events))
	}
	if second.Events[0].Type != "command-start" || second.Events[0].Command != "codex" {
		t.Fatalf("Event = %#v, want command-start codex", second.Events[0])
	}
}

func TestCommandStateOutputFilterPreservesSplitBracketedWindowsTextFallbackPrefix(t *testing.T) {
	filter := newCommandStateOutputFilter()

	first := filter.Filter("before ]777;tui-")
	second := filter.Filter("helper;command-start;Y2FsYw==\aafter")

	if first.Data != "before ]777;tui-" {
		t.Fatalf("first Data = %q, want before ]777;tui-", first.Data)
	}
	if len(first.Events) != 0 {
		t.Fatalf("first Events = %#v, want none", first.Events)
	}
	if second.Data != "helper;command-start;Y2FsYw==\aafter" {
		t.Fatalf("second Data = %q, want helper;command-start;Y2FsYw==\\aafter", second.Data)
	}
	if len(second.Events) != 0 {
		t.Fatalf("second Events = %#v, want none", second.Events)
	}
}

func TestCommandStateOutputFilterDropsInvalidPrivatePayload(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("before\x1b]777;todoai;command-start;not-base64\aafter")

	if result.Data != "beforeafter" {
		t.Fatalf("Data = %q, want beforeafter", result.Data)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterDropsInvalidWindowsTextPayload(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("before 777;todoai;command-start;not-base64\r\nafter")

	if result.Data != "before after" {
		t.Fatalf("Data = %q, want before after", result.Data)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterDropsInvalidWindowsTextPayloadWithoutLineTerminator(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("before 777;todoai;command-start;not-base64 after")

	if result.Data != "before  after" {
		t.Fatalf("Data = %q, want before  after", result.Data)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterPreservesNonApplicationOutput(t *testing.T) {
	filter := newCommandStateOutputFilter()
	input := "before\x1b]999;other;payload\a after 777;not-helper;command-start;Y29kZXg=\r\n after 777;todone;command-start;Y29kZXg=\r\n"

	result := filter.Filter(input)

	if result.Data != input {
		t.Fatalf("Data = %q, want original input", result.Data)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterPreservesOrdinaryBase64LikeOutput(t *testing.T) {
	filter := newCommandStateOutputFilter()
	input := "token Y2FsYw== still visible\r\n"

	result := filter.Filter(input)

	if result.Data != input {
		t.Fatalf("Data = %q, want %q", result.Data, input)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterDoesNotDelayOrdinaryTrailingSeven(t *testing.T) {
	filter := newCommandStateOutputFilter()

	result := filter.Filter("version 7")

	if result.Data != "version 7" {
		t.Fatalf("Data = %q, want version 7", result.Data)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}

func TestCommandStateOutputFilterPreservesLongerCommandEndText(t *testing.T) {
	filter := newCommandStateOutputFilter()
	input := "777;tui-helper;command-ending"

	result := filter.Filter(input)

	if result.Data != input {
		t.Fatalf("Data = %q, want %q", result.Data, input)
	}
	if len(result.Events) != 0 {
		t.Fatalf("Events = %#v, want none", result.Events)
	}
}
