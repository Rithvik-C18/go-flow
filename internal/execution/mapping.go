package execution

import (
	"fmt"
	"strings"
)

func applyInputMapping(config map[string]any, ctx *ExecutionContext) error {
	raw, ok := config["input"]
	if !ok {
		return nil
	}

	mapping, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("input mapping must be an object")
	}

	newInput := make(map[string]any, len(mapping))

	for outName, src := range mapping {
		value, err := resolveValue(src, ctx.Input, nil, ctx.Params)
		if err != nil {
			return fmt.Errorf("input mapping for %q: %w", outName, err)
		}
		newInput[outName] = value
	}

	ctx.Input = newInput

	return nil
}

func applyOutputMapping(config map[string]any, ctx *ExecutionContext) error {
	raw, ok := config["output"]
	if !ok {
		return nil
	}

	mapping, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("output mapping must be an object")
	}

	newOutput := make(map[string]any, len(mapping))

	for outName, src := range mapping {
		value, err := resolveValue(src, ctx.Output, ctx.Input, ctx.Params)
		if err != nil {
			return fmt.Errorf("output mapping for %q: %w", outName, err)
		}
		newOutput[outName] = value
	}

	ctx.Output = newOutput

	return nil
}

func resolveValue(src any, primary, secondary, params map[string]any) (any, error) {
	str, isString := src.(string)
	if !isString {
		return src, nil
	}

	key, isTemplate := parseTemplate(str)
	if !isTemplate {
		return str, nil
	}

	data, key, ok := resolveTarget(key, primary, secondary, params)
	if !ok {
		return nil, fmt.Errorf("key %q not available in this scope (available: input.*, params.*, or direct field)", key)
	}

	value, found, err := lookupValue(data, key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("key %q not found in data (resolved target: %s)", key, key)
	}

	return value, nil
}

func resolveOptional(src any, primary, secondary, params map[string]any) (any, bool, error) {
	str, isString := src.(string)
	if !isString {
		return src, true, nil
	}

	key, isTemplate := parseTemplate(str)
	if !isTemplate {
		return str, true, nil
	}

	data, key, ok := resolveTarget(key, primary, secondary, params)
	if !ok {
		return nil, false, nil
	}

	value, found, err := lookupValue(data, key)
	if err != nil {
		return nil, false, fmt.Errorf("error looking up key %q: %w", key, err)
	}

	return value, found, nil
}

func resolveTarget(key string, primary, secondary, params map[string]any) (map[string]any, string, bool) {
	var data map[string]any

	switch {
	case strings.HasPrefix(key, "input."):
		if secondary == nil {
			return nil, "", false
		}
		data = secondary
		key = strings.TrimPrefix(key, "input.")
	case strings.HasPrefix(key, "params."):
		if params == nil {
			return nil, "", false
		}
		data = params
		key = strings.TrimPrefix(key, "params.")
	default:
		data = primary
	}

	return data, key, true
}

func parseTemplate(str string) (string, bool) {
	if !strings.HasPrefix(str, "{{") || !strings.HasSuffix(str, "}}") {
		return "", false
	}
	return strings.TrimSpace(str[2 : len(str)-2]), true
}

func lookupValue(data map[string]any, path string) (any, bool, error) {
	parts := strings.Split(path, ".")

	var current any = data

	for _, part := range parts {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("cannot traverse %q: value is not an object", part)
		}

		current, ok = obj[part]
		if !ok {
			return nil, false, nil
		}
	}

	return current, true, nil
}
