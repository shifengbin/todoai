package main

import (
	"encoding/base64"
	"strings"
)

const (
	commandStateRawPrefix  = "\x1b]777;tui-helper;"
	commandStateTextPrefix = "777;tui-helper;"
	commandStateMaxPending = 8192
)

type TerminalCommandStateEvent struct {
	ProjectID     string `json:"projectId"`
	TodoID        string `json:"todoId,omitempty"`
	TodoProjectID string `json:"todoProjectId,omitempty"`
	TerminalID    string `json:"terminalId"`
	Type          string `json:"type"`
	Command       string `json:"command,omitempty"`
}

type commandStateOutputFilter struct {
	pending string
}

type commandStateFilterResult struct {
	Data   string
	Events []TerminalCommandStateEvent
}

func newCommandStateOutputFilter() *commandStateOutputFilter {
	return &commandStateOutputFilter{}
}

func (filter *commandStateOutputFilter) Filter(data string) commandStateFilterResult {
	input := filter.pending + data
	filter.pending = ""

	var result commandStateFilterResult
	var output strings.Builder

	for len(input) > 0 {
		rawIndex := strings.Index(input, commandStateRawPrefix)
		textIndex := strings.Index(input, commandStateTextPrefix)
		index, raw := nextCommandStateIndex(rawIndex, textIndex)
		if index == -1 {
			emit, pending := splitCommandStatePendingPrefix(input)
			output.WriteString(emit)
			filter.pending = pending
			break
		}

		output.WriteString(input[:index])
		var parsed commandStateParseResult
		if raw {
			parsed = parseRawCommandState(input[index:])
		} else {
			parsed = parseTextCommandState(input[index:])
		}

		if parsed.needsMore {
			filter.pending = input[index:]
			if len(filter.pending) > commandStateMaxPending {
				output.WriteString(filter.pending)
				filter.pending = ""
			}
			break
		}
		if !parsed.consumed {
			output.WriteByte(input[index])
			input = input[index+1:]
			continue
		}
		if parsed.event != nil {
			result.Events = append(result.Events, *parsed.event)
		}
		input = input[index+parsed.consumedBytes:]
	}

	result.Data = output.String()
	return result
}

type commandStateParseResult struct {
	consumed      bool
	needsMore     bool
	consumedBytes int
	event         *TerminalCommandStateEvent
}

func nextCommandStateIndex(rawIndex int, textIndex int) (int, bool) {
	if rawIndex == -1 && textIndex == -1 {
		return -1, false
	}
	if rawIndex != -1 && (textIndex == -1 || rawIndex <= textIndex) {
		return rawIndex, true
	}
	return textIndex, false
}

func parseRawCommandState(input string) commandStateParseResult {
	terminatorIndex, terminatorSize := rawCommandStateTerminator(input)
	if terminatorIndex == -1 {
		return commandStateParseResult{needsMore: true}
	}

	payload := input[len(commandStateRawPrefix):terminatorIndex]
	event, consume := parseCommandStatePayload(payload)
	if !consume {
		return commandStateParseResult{}
	}
	return commandStateParseResult{
		consumed:      true,
		consumedBytes: terminatorIndex + terminatorSize,
		event:         event,
	}
}

func rawCommandStateTerminator(input string) (int, int) {
	belIndex := strings.IndexByte(input, '\a')
	stIndex := strings.Index(input, "\x1b\\")
	if belIndex == -1 {
		if stIndex == -1 {
			return -1, 0
		}
		return stIndex, 2
	}
	if stIndex == -1 || belIndex < stIndex {
		return belIndex, 1
	}
	return stIndex, 2
}

func parseTextCommandState(input string) commandStateParseResult {
	payload := input[len(commandStateTextPrefix):]
	if strings.HasPrefix(payload, "command-end") {
		consumed := len(commandStateTextPrefix) + len("command-end")
		terminatorSize := textLineTerminatorSize(input[consumed:])
		if consumed < len(input) && terminatorSize == 0 {
			return commandStateParseResult{}
		}
		return commandStateParseResult{
			consumed:      true,
			consumedBytes: consumed + terminatorSize,
			event:         &TerminalCommandStateEvent{Type: "command-end"},
		}
	}

	const commandStartPrefix = "command-start;"
	if !strings.HasPrefix(payload, commandStartPrefix) {
		if len(payload) < len(commandStartPrefix) && strings.HasPrefix(commandStartPrefix, payload) {
			return commandStateParseResult{needsMore: true}
		}
		return commandStateParseResult{}
	}

	payloadStart := len(commandStateTextPrefix) + len(commandStartPrefix)
	payloadEnd, consumedBytes, complete := textCommandStartPayloadBounds(input, payloadStart)
	if !complete {
		return commandStateParseResult{needsMore: true}
	}
	if payloadEnd == payloadStart || payloadEnd == len(input) {
		return commandStateParseResult{needsMore: true}
	}

	event, consume := parseCommandStatePayload("command-start;" + input[payloadStart:payloadEnd])
	if !consume {
		return commandStateParseResult{}
	}
	return commandStateParseResult{
		consumed:      true,
		consumedBytes: consumedBytes,
		event:         event,
	}
}

func textCommandStartPayloadBounds(input string, payloadStart int) (int, int, bool) {
	if terminatorIndex, terminatorSize := textTerminatorIndex(input[payloadStart:]); terminatorIndex != -1 {
		payloadEnd := payloadStart + terminatorIndex
		return payloadEnd, payloadEnd + terminatorSize, true
	}

	payloadEnd := payloadStart
	for payloadEnd < len(input) && isBase64Byte(input[payloadEnd]) {
		payloadEnd++
	}
	if payloadEnd == len(input) {
		return 0, 0, false
	}
	if payloadEnd < len(input) && !isTextPayloadSeparator(input[payloadEnd]) {
		for payloadEnd < len(input) && !isTextPayloadSeparator(input[payloadEnd]) {
			payloadEnd++
		}
	}
	return payloadEnd, payloadEnd, true
}

func textTerminatorIndex(input string) (int, int) {
	for index := 0; index < len(input); index++ {
		switch input[index] {
		case '\r':
			if index+1 < len(input) && input[index+1] == '\n' {
				return index, 2
			}
			return index, 1
		case '\n':
			return index, 1
		}
	}
	return -1, 0
}

func textLineTerminatorSize(input string) int {
	if strings.HasPrefix(input, "\r\n") {
		return 2
	}
	if strings.HasPrefix(input, "\r") || strings.HasPrefix(input, "\n") {
		return 1
	}
	return 0
}

func isBase64Byte(value byte) bool {
	return value >= 'A' && value <= 'Z' ||
		value >= 'a' && value <= 'z' ||
		value >= '0' && value <= '9' ||
		value == '+' ||
		value == '/' ||
		value == '='
}

func isTextPayloadSeparator(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n'
}

func parseCommandStatePayload(payload string) (*TerminalCommandStateEvent, bool) {
	if payload == "command-end" {
		return &TerminalCommandStateEvent{Type: "command-end"}, true
	}
	if strings.HasPrefix(payload, "command-start;") {
		encoded := strings.TrimPrefix(payload, "command-start;")
		command, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, true
		}
		return &TerminalCommandStateEvent{Type: "command-start", Command: string(command)}, true
	}
	return nil, false
}

func splitCommandStatePendingPrefix(input string) (string, string) {
	maxLength := len(commandStateRawPrefix)
	if len(commandStateTextPrefix) > maxLength {
		maxLength = len(commandStateTextPrefix)
	}
	if len(input) < maxLength {
		maxLength = len(input)
	}
	for length := maxLength; length > 0; length-- {
		suffix := input[len(input)-length:]
		if strings.HasPrefix(commandStateRawPrefix, suffix) {
			return input[:len(input)-length], suffix
		}
	}
	return input, ""
}
