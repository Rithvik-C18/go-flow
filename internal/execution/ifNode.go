package execution

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Condition struct {
	left     any
	operator string
	right    any
}

type IfNode struct {
	combine    string
	conditions []Condition
}

func NewIfNode(config map[string]any) (*IfNode, error) {
	combine := "and"
	if v, ok := config["combine"].(string); ok && v != "" {
		if v != "and" && v != "or" {
			return nil, fmt.Errorf("if node: combine must be \"and\" or \"or\"")
		}
		combine = v
	}

	node := &IfNode{combine: combine}

	var raw []any
	if v, ok := config["conditions"]; ok {
		switch r := v.(type) {
		case []any:
			raw = r
		case []map[string]any:
			for _, m := range r {
				raw = append(raw, m)
			}
		default:
			return nil, fmt.Errorf("if node: conditions must be an array")
		}
	}

	if raw == nil {
		cond, err := parseCondition(config)
		if err != nil {
			return nil, fmt.Errorf("if node: %w", err)
		}
		node.conditions = []Condition{cond}
		return node, nil
	}

	for i, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("if node: condition %d must be an object", i)
		}

		cond, err := parseCondition(m)
		if err != nil {
			return nil, fmt.Errorf("if node: condition %d: %w", i, err)
		}

		node.conditions = append(node.conditions, cond)
	}

	if len(node.conditions) == 0 {
		return nil, fmt.Errorf("if node: at least one condition is required")
	}

	return node, nil
}

func parseCondition(m map[string]any) (Condition, error) {
	left, ok := m["left"]
	if !ok {
		return Condition{}, fmt.Errorf("missing left operand")
	}

	operator, ok := m["operator"].(string)
	if !ok || operator == "" {
		return Condition{}, fmt.Errorf("missing or invalid operator")
	}

	cond := Condition{left: left, operator: operator}

	switch operator {
	case "exists", "notExists", "isEmpty", "notEmpty", "isTruthy", "isFalsy":
	case "==", "!=", ">", "<", ">=", "<=", "contains", "notContains", "startsWith", "endsWith":
		right, ok := m["right"]
		if !ok {
			return Condition{}, fmt.Errorf("operator %q requires a right operand", operator)
		}
		cond.right = right
	default:
		return Condition{}, fmt.Errorf("unsupported operator %q", operator)
	}

	return cond, nil
}

func (n *IfNode) Execute(ctx *ExecutionContext) error {
	result := n.combine == "and"

	for _, cond := range n.conditions {
		ok, err := cond.evaluate(ctx)
		if err != nil {
			return err
		}

		if n.combine == "and" {
			result = result && ok
			if !result {
				break
			}
		} else {
			result = result || ok
			if result {
				break
			}
		}
	}

	ctx.Output = map[string]any{"result": result}

	return nil
}

func (c *Condition) evaluate(ctx *ExecutionContext) (bool, error) {
	left, found, err := resolveOptional(c.left, ctx.Input, nil, ctx.Params)
	if err != nil {
		return false, fmt.Errorf("if node: error resolving left operand: %w", err)
	}

	switch c.operator {
	case "exists":
		return found, nil
	case "notExists":
		return !found, nil
	case "isEmpty":
		if !found {
			return true, nil
		}
		return isEmpty(left), nil
	case "notEmpty":
		if !found {
			return false, nil
		}
		return !isEmpty(left), nil
	case "isTruthy":
		if !found {
			return false, nil
		}
		return isTruthy(left), nil
	case "isFalsy":
		if !found {
			return true, nil
		}
		return !isTruthy(left), nil
	}

	right, err := resolveValue(c.right, ctx.Input, nil, ctx.Params)
	if err != nil {
		return false, fmt.Errorf("if node: error resolving right operand: %w", err)
	}

	switch c.operator {
	case "contains":
		return contains(left, right)
	case "notContains":
		ok, err := contains(left, right)
		return !ok, err
	case "startsWith", "endsWith":
		s, ok := left.(string)
		if !ok {
			return false, fmt.Errorf("if node: %q requires a string left operand, got %T (value: %v)", c.operator, left, left)
		}
		e, ok := right.(string)
		if !ok {
			return false, fmt.Errorf("if node: %q requires a string right operand, got %T (value: %v)", c.operator, right, right)
		}
		if c.operator == "startsWith" {
			return strings.HasPrefix(s, e), nil
		}
		return strings.HasSuffix(s, e), nil
	case "==", "!=", ">", "<", ">=", "<=":
		return compare(c.operator, left, right)
	}

	return false, fmt.Errorf("if node: unsupported operator %q", c.operator)
}

func compare(operator string, left, right any) (bool, error) {
	// Handle nil values
	if left == nil || right == nil {
		switch operator {
		case "==":
			return left == nil && right == nil, nil
		case "!=":
			return left != nil || right != nil, nil
		case ">":
			return left != nil && right == nil, nil
		case "<":
			return left == nil && right != nil, nil
		case ">=":
			return left == right || (left != nil && right == nil), nil
		case "<=":
			return left == right || (left == nil && right != nil), nil
		default:
			return false, fmt.Errorf("if node: operator %q with nil values only supports ==, !=, >, <, >=, <=", operator)
		}
	}

	if operator == "==" {
		return equal(left, right), nil
	}
	if operator == "!=" {
		return !equal(left, right), nil
	}

	leftNum, leftIsNum := toFloat64(left)
	rightNum, rightIsNum := toFloat64(right)
	if leftIsNum && rightIsNum {
		switch operator {
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)
	if leftIsStr && rightIsStr {
		cmp := strings.Compare(leftStr, rightStr)
		switch operator {
		case ">":
			return cmp > 0, nil
		case "<":
			return cmp < 0, nil
		case ">=":
			return cmp >= 0, nil
		case "<=":
			return cmp <= 0, nil
		}
	}

	return false, fmt.Errorf("if node: operator %q requires numeric or string operands, got %T (value: %v) and %T (value: %v)", operator, left, left, right, right)
}

func equal(left, right any) bool {
	leftNum, okLeft := toFloat64(left)
	rightNum, okRight := toFloat64(right)
	if okLeft && okRight {
		return leftNum == rightNum
	}
	return reflect.DeepEqual(left, right)
}

func contains(container, element any) (bool, error) {
	if container == nil {
		return false, nil
	}

	switch c := container.(type) {
	case string:
		e, ok := element.(string)
		if !ok {
			return false, fmt.Errorf("if node: contains on a string requires a string right operand")
		}
		return strings.Contains(c, e), nil
	case []any:
		for _, item := range c {
			if equal(item, element) {
				return true, nil
			}
		}
		return false, nil
	case map[string]any:
		key, ok := element.(string)
		if !ok {
			return false, fmt.Errorf("if node: contains on an object requires a string key")
		}
		_, ok = c[key]
		return ok, nil
	default:
		return false, fmt.Errorf("if node: contains requires a string, array or object, got %T", container)
	}
}

func isEmpty(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	case []any:
		return len(t) == 0
	case map[string]any:
		return len(t) == 0
	default:
		return false
	}
}

func isTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int() != 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return rv.Uint() != 0
		case reflect.Float32:
			return rv.Float() != 0
		case reflect.Array, reflect.Slice:
			return rv.Len() > 0
		case reflect.Map:
			return rv.Len() > 0
		}
		return true
	}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int()), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), true
		}
		return 0, false
	}
}
