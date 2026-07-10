package main

import (
	"errors"
	"regexp"
	"strings"
)

var todoLifecycleScriptParameterNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type TodoLifecycleScriptParameter struct {
	Name         string `json:"name"`
	Label        string `json:"label,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Required     bool   `json:"required,omitempty"`
}

func normalizeTodoLifecycleScriptParameters(parameters []TodoLifecycleScriptParameter) ([]TodoLifecycleScriptParameter, error) {
	if len(parameters) == 0 {
		return nil, nil
	}
	normalized := make([]TodoLifecycleScriptParameter, 0, len(parameters))
	seen := map[string]bool{}
	for _, parameter := range parameters {
		name := strings.TrimSpace(parameter.Name)
		if name == "" {
			return nil, errors.New("lifecycle script parameter name is required")
		}
		if !todoLifecycleScriptParameterNamePattern.MatchString(name) {
			return nil, errors.New("lifecycle script parameter name is invalid")
		}
		if reservedTodoLifecycleScriptParameterName(name) {
			return nil, errors.New("lifecycle script parameter name is reserved")
		}
		if seen[name] {
			return nil, errors.New("lifecycle script parameter name is duplicated")
		}
		seen[name] = true
		normalized = append(normalized, TodoLifecycleScriptParameter{
			Name:         name,
			Label:        strings.TrimSpace(parameter.Label),
			Description:  strings.TrimSpace(parameter.Description),
			DefaultValue: parameter.DefaultValue,
			Required:     parameter.Required,
		})
	}
	return normalized, nil
}

func reservedTodoLifecycleScriptParameterName(name string) bool {
	switch name {
	case "__proto__", "prototype", "constructor":
		return true
	default:
		return false
	}
}

func normalizeTodoLifecycleScriptParameterValues(parameters []TodoLifecycleScriptParameter, values map[string]string) (map[string]string, error) {
	if len(parameters) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(parameters))
	for _, parameter := range parameters {
		value, ok := values[parameter.Name]
		if !ok {
			value = parameter.DefaultValue
		}
		if parameter.Required && strings.TrimSpace(value) == "" {
			return nil, errors.New("lifecycle script parameter value is required")
		}
		normalized[parameter.Name] = value
	}
	return normalized, nil
}
