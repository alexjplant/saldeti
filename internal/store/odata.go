package store

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/saldeti/saldeti/internal/model"
)

// ApplyOData applies OData query options to a slice of items
func ApplyOData[T any](items []T, opts model.ListOptions) ([]T, int, error) {
	// Convert items to maps for filtering
	maps := make([]map[string]interface{}, len(items))
	for i, item := range items {
		m, err := structToMap(item)
		if err != nil {
			return nil, 0, err
		}
		maps[i] = m
	}

	// Apply search if specified
	searchedMaps := maps
	if opts.Search != "" {
		searchedMaps = applySearch(maps, opts.Search)
	}

	// Apply filter if specified
	filteredMaps := searchedMaps
	if opts.Filter != "" {
		var err error
		filteredMaps, err = applyFilter(searchedMaps, opts.Filter)
		if err != nil {
			return nil, 0, err
		}
	}

	// Apply sorting if specified
	if opts.OrderBy != "" {
		applySorting(filteredMaps, opts.OrderBy)
	}

	// Calculate total count before pagination
	totalCount := len(filteredMaps)

	// Apply pagination
	start := opts.Skip
	if start < 0 {
		start = 0
	}
	end := len(filteredMaps)
	if opts.Top > 0 && start+opts.Top < end {
		end = start + opts.Top
	}
	if start > end {
		start = end
	}
	paginatedMaps := filteredMaps[start:end]

	// Apply field selection if specified
	if len(opts.Select) > 0 {
		for i, m := range paginatedMaps {
			paginatedMaps[i] = selectFields(m, opts.Select)
		}
	}

	// Convert maps back to structs
	result := make([]T, len(paginatedMaps))
	for i, m := range paginatedMaps {
		item, err := mapToStruct[T](m)
		if err != nil {
			return nil, 0, err
		}
		result[i] = item
	}

	return result, totalCount, nil
}

// structToMap converts a struct to a map[string]interface{}
func structToMap(item interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get JSON tag name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty suffix
		jsonName := strings.Split(jsonTag, ",")[0]

		// Handle zero values for pointers
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				result[jsonName] = nil
			} else {
				result[jsonName] = value.Elem().Interface()
			}
		} else {
			result[jsonName] = value.Interface()
		}
	}

	return result, nil
}

// mapToStruct converts a map back to a struct
func mapToStruct[T any](m map[string]interface{}) (T, error) {
	var result T
	v := reflect.ValueOf(&result).Elem()
	t := reflect.TypeOf(result)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		if val, ok := m[jsonName]; ok {
			fieldValue := v.Field(i)

			if fieldValue.Kind() == reflect.Ptr {
				if val == nil {
					fieldValue.Set(reflect.Zero(fieldValue.Type()))
				} else {
					ptr := reflect.New(fieldValue.Type().Elem())
					ptr.Elem().Set(reflect.ValueOf(val))
					fieldValue.Set(ptr)
				}
			} else {
				if val != nil {
					fieldValue.Set(reflect.ValueOf(val))
				}
			}
		}
	}

	return result, nil
}

// applyFilter applies OData filter expression to items
func applyFilter(items []map[string]interface{}, filter string) ([]map[string]interface{}, error) {
	if filter == "" {
		return items, nil
	}

	// Parse the filter expression
	expr, err := parseFilterExpression(filter)
	if err != nil {
		return nil, err
	}

	// Evaluate filter for each item
	result := make([]map[string]interface{}, 0)
	for _, item := range items {
		matches, err := evaluateExpression(expr, item)
		if err != nil {
			return nil, err
		}
		if matches {
			result = append(result, item)
		}
	}

	return result, nil
}

// parseFilterExpression parses OData filter expression into an AST
func parseFilterExpression(filter string) (*filterNode, error) {
	// Remove whitespace
	filter = strings.TrimSpace(filter)

	// Handle parentheses - check if entire expression is wrapped in parentheses
	if strings.HasPrefix(filter, "(") {
		parenCount := 0
		for i, ch := range filter {
			if ch == '(' {
				parenCount++
			} else if ch == ')' {
				parenCount--
				if parenCount == 0 {
					// Found matching closing paren
					if i == len(filter)-1 {
						// Entire expression is wrapped in parentheses
						return parseFilterExpression(filter[1 : len(filter)-1])
					}
					break
				}
			}
		}
	}

	// Try to parse as logical expression
	if node := parseLogicalExpression(filter); node != nil {
		return node, nil
	}

	// Try to parse as comparison expression
	if node := parseComparisonExpression(filter); node != nil {
		return node, nil
	}

	// Try to parse as function call
	if node := parseFunctionCall(filter); node != nil {
		return node, nil
	}

	return nil, fmt.Errorf("unable to parse filter expression: %s", filter)
}

type filterNode struct {
	operator   string
	left       *filterNode
	right      *filterNode
	value      interface{}
	property   string
	function   string
	args       []*filterNode
	nestedPath string // for any() with nested property access, e.g., "skuId" from "a/skuId"
}

// parseLogicalExpression parses logical operators (and, or)
func parseLogicalExpression(expr string) *filterNode {
	// Use proper operator precedence: and binds tighter than or
	return parseOr(expr)
}

// parseOr scans for "or" at top level FIRST, then parses each side.
// This ensures logical operators take precedence over greedy comparison regex.
func parseOr(expr string) *filterNode {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Scan for " or " at the top level (outside parentheses and quotes) FIRST
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if ch == '(' {
			// Skip to matching closing paren
			depth := 1
			for j := i + 1; j < len(expr) && depth > 0; j++ {
				if expr[j] == '(' {
					depth++
				} else if expr[j] == ')' {
					depth--
				}
			}
			// Advance i past the closing paren
			for j := i + 1; j < len(expr); j++ {
				if expr[j] == ')' {
					i = j
					break
				}
			}
		} else if ch == '\'' {
			// Skip to matching close quote
			for j := i + 1; j < len(expr); j++ {
				if expr[j] == '\'' {
					i = j
					break
				}
			}
		} else if i+4 <= len(expr) && strings.EqualFold(expr[i:i+4], " or ") {
			// Found OR at top level — split and parse each side
			leftStr := strings.TrimSpace(expr[:i])
			rightStr := strings.TrimSpace(expr[i+4:])
			left := parseAnd(leftStr)
			right := parseAnd(rightStr)
			if left != nil && right != nil {
				return &filterNode{
					operator: "or",
					left:     left,
					right:    right,
				}
			}
		}
	}

	// No "or" found at top level, delegate to parseAnd
	return parseAnd(expr)
}

// parseAnd scans for "and" at top level FIRST, then parses each side.
func parseAnd(expr string) *filterNode {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Scan for " and " at the top level (outside parentheses and quotes) FIRST
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if ch == '(' {
			// Skip to matching closing paren
			depth := 1
			for j := i + 1; j < len(expr) && depth > 0; j++ {
				if expr[j] == '(' {
					depth++
				} else if expr[j] == ')' {
					depth--
				}
			}
			for j := i + 1; j < len(expr); j++ {
				if expr[j] == ')' {
					i = j
					break
				}
			}
		} else if ch == '\'' {
			// Skip to matching close quote
			for j := i + 1; j < len(expr); j++ {
				if expr[j] == '\'' {
					i = j
					break
				}
			}
		} else if i+5 <= len(expr) && strings.EqualFold(expr[i:i+5], " and ") {
			// Found AND at top level — split and parse each side
			leftStr := strings.TrimSpace(expr[:i])
			rightStr := strings.TrimSpace(expr[i+5:])
			left := parsePrimary(leftStr)
			right := parsePrimary(rightStr)
			if left != nil && right != nil {
				return &filterNode{
					operator: "and",
					left:     left,
					right:    right,
				}
			}
		}
	}

	// No "and" found at top level, delegate to parsePrimary
	return parsePrimary(expr)
}

// parsePrimary parses a primary expression (comparison, function call, or parenthesized expression)
func parsePrimary(expr string) *filterNode {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Handle parentheses - check if entire expression is wrapped in parentheses
	if strings.HasPrefix(expr, "(") {
		parenCount := 0
		for i, ch := range expr {
			if ch == '(' {
				parenCount++
			} else if ch == ')' {
				parenCount--
				if parenCount == 0 {
					// Found matching closing paren
					if i == len(expr)-1 {
						// Entire expression is wrapped in parentheses
						return parseOr(expr[1 : len(expr)-1])
					}
					break
				}
			}
		}
	}

	// Try to parse as function call
	if node := parseFunctionCall(expr); node != nil {
		return node
	}

	// Try to parse as comparison expression
	if node := parseComparisonExpression(expr); node != nil {
		return node
	}

	return nil
}

// parseComparisonExpression parses comparison operators (eq, ne, gt, ge, lt, le)
func parseComparisonExpression(expr string) *filterNode {
	// Regex for comparison operators (case-insensitive)
	pattern := `^(.+?)\s+(eq|ne|gt|ge|lt|le)\s+(.+)$`
	re := regexp.MustCompile("(?i)" + pattern)
	matches := re.FindStringSubmatch(expr)

	if matches == nil {
		return nil
	}

	property := strings.TrimSpace(matches[1])
	operator := strings.ToLower(strings.TrimSpace(matches[2]))
	valueStr := strings.TrimSpace(matches[3])

	// Parse value (case-insensitive for boolean and null)
	var value interface{}
	if strings.EqualFold(valueStr, "null") {
		value = nil
	} else if strings.EqualFold(valueStr, "true") {
		value = true
	} else if strings.EqualFold(valueStr, "false") {
		value = false
	} else if strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'") {
		// String value
		value = valueStr[1 : len(valueStr)-1]
	} else if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
		// Numeric value
		value = num
	} else {
		// Could be another property or unquoted string
		value = valueStr
	}

	return &filterNode{
		operator: operator,
		property: property,
		value:    value,
	}
}

// parseFunctionCall parses function calls (startswith, endswith, contains, any)
func parseFunctionCall(expr string) *filterNode {
	// Check if it's an any() expression: property/any(a:a eq 'value')
	if strings.Contains(expr, "/any(") {
		return parseAnyFunction(expr)
	}

	// Parse regular function calls: function(property, 'value')
	// Find the opening parenthesis
	openParen := strings.Index(expr, "(")
	if openParen == -1 {
		return nil
	}

	function := strings.ToLower(strings.TrimSpace(expr[:openParen]))
	argsStr := strings.TrimSpace(expr[openParen+1:])

	// Find matching closing parenthesis
	if len(argsStr) == 0 || argsStr[len(argsStr)-1] != ')' {
		return nil
	}
	argsStr = argsStr[:len(argsStr)-1]

	// Parse arguments procedurally
	property, value, err := parseFunctionArgs(argsStr)
	if err != nil {
		return nil
	}

	// Validate function
	validFunctions := map[string]bool{
		"startswith": true,
		"endswith":   true,
		"contains":   true,
	}

	if !validFunctions[function] {
		return nil
	}

	return &filterNode{
		function: function,
		property: property,
		value:    value,
	}
}

// parseFunctionArgs parses function arguments procedurally
func parseFunctionArgs(argsStr string) (property string, value string, err error) {
	argsStr = strings.TrimSpace(argsStr)

	// Find the comma that separates property and value
	// Need to track quotes to skip commas inside quoted strings
	inQuote := false
	commaPos := -1
	for i := 0; i < len(argsStr); i++ {
		ch := argsStr[i]
		if ch == '\'' {
			inQuote = !inQuote
		} else if ch == ',' && !inQuote {
			commaPos = i
			break
		}
	}

	if commaPos == -1 {
		return "", "", fmt.Errorf("missing comma in function arguments")
	}

	property = strings.TrimSpace(argsStr[:commaPos])
	valueStr := strings.TrimSpace(argsStr[commaPos+1:])

	// Parse the value (should be a quoted string)
	valueStr = strings.TrimSpace(valueStr)
	if len(valueStr) < 2 || valueStr[0] != '\'' || valueStr[len(valueStr)-1] != '\'' {
		return "", "", fmt.Errorf("value must be a quoted string")
	}

	value = valueStr[1 : len(valueStr)-1]

	// Unescape doubled quotes ('' -> ')
	value = strings.ReplaceAll(value, "''", "'")

	return property, value, nil
}

// parseAnyFunction parses any() function expressions
// Supports both flat and nested lambda paths:
//   - property/any(a:a eq 'value')               — flat
//   - property/any(a:a/nestedProp eq 'value')     — nested
func parseAnyFunction(expr string) *filterNode {
	// Format: property/any(a:a eq 'value') or property/any(a:a/nested eq 'value')
	parts := strings.SplitN(expr, "/any(", 2)
	if len(parts) != 2 {
		return nil
	}

	property := strings.TrimSpace(parts[0])
	innerExpr := parts[1]

	// Remove closing parenthesis
	if len(innerExpr) == 0 || innerExpr[len(innerExpr)-1] != ')' {
		return nil
	}
	innerExpr = innerExpr[:len(innerExpr)-1]

	// Extract the lambda variable and condition
	// Format: a:a eq 'value' or a:a/nestedProp eq 'value'
	colonPos := strings.Index(innerExpr, ":")
	if colonPos == -1 {
		return nil
	}

	lambdaVar := strings.TrimSpace(innerExpr[:colonPos])
	condition := strings.TrimSpace(innerExpr[colonPos+1:])

	// Parse the inner comparison expression
	comparisonNode := parseComparisonExpression(condition)
	if comparisonNode != nil {
		// Check if the property contains a nested path (e.g., "a/skuId")
		nodeProperty := comparisonNode.property
		var nestedPath string
		if strings.HasPrefix(nodeProperty, lambdaVar+"/") {
			nestedPath = nodeProperty[len(lambdaVar)+1:]
		}

		return &filterNode{
			function:   "any",
			property:   property,
			value:      comparisonNode.value,
			operator:   comparisonNode.operator,
			nestedPath: nestedPath,
		}
	}

	return nil
}

// evaluateExpression evaluates a filter node against an item
func evaluateExpression(node *filterNode, item map[string]interface{}) (bool, error) {
	if node == nil {
		return true, nil
	}

	// Handle logical operators
	switch node.operator {
	case "and":
		left, err := evaluateExpression(node.left, item)
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil
		}
		right, err := evaluateExpression(node.right, item)
		if err != nil {
			return false, err
		}
		return left && right, nil
	case "or":
		left, err := evaluateExpression(node.left, item)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		right, err := evaluateExpression(node.right, item)
		if err != nil {
			return false, err
		}
		return left || right, nil
	}

	// Handle function calls FIRST (before comparison operators)
	// any() filterNodes have both operator and function set; function takes precedence
	if node.function != "" {
		itemValue, exists := item[node.property]
		if !exists {
			// Property doesn't exist in item
			return false, nil
		}

		// Handle any() function for array properties
		if node.function == "any" {
			// itemValue should be a slice for any() to work
			switch v := itemValue.(type) {
			case []string:
				// Flat string array: groupTypes/any(a:a eq 'Unified')
				for _, elem := range v {
					match, err := compareValues(elem, node.value, node.operator)
					if err == nil && match {
						return true, nil
					}
				}
				return false, nil
			case []interface{}:
				// Mixed array
				for _, elem := range v {
					match, err := compareValues(elem, node.value, node.operator)
					if err == nil && match {
						return true, nil
					}
				}
				return false, nil
			default:
				// Could be a typed slice of structs (e.g., []AssignedLicense)
				// Use reflection to iterate and handle nested paths
				rv := reflect.ValueOf(itemValue)
				if rv.Kind() != reflect.Slice {
					return false, nil
				}
				for i := 0; i < rv.Len(); i++ {
					elem := rv.Index(i)
					// Dereference pointer if needed
					if elem.Kind() == reflect.Ptr {
						elem = elem.Elem()
					}
					if elem.Kind() != reflect.Struct {
						// Not a struct, try direct comparison
						match, err := compareValues(elem.Interface(), node.value, node.operator)
						if err == nil && match {
							return true, nil
						}
						continue
					}

					if node.nestedPath != "" {
						// Nested property access: e.g., a/skuId
						propValue := getNestedFieldValue(elem, node.nestedPath)
						match, err := compareValues(propValue, node.value, node.operator)
						if err == nil && match {
							return true, nil
						}
					} else {
						// Flat comparison on struct - compare struct itself
						match, err := compareValues(elem.Interface(), node.value, node.operator)
						if err == nil && match {
							return true, nil
						}
					}
				}
				return false, nil
			}
		}

		// Handle string functions
		strValue, ok := itemValue.(string)
		if !ok {
			// Property is not a string
			return false, nil
		}

		funcValue, ok := node.value.(string)
		if !ok {
			return false, fmt.Errorf("function value must be string")
		}

		switch node.function {
		case "startswith":
			return strings.HasPrefix(strValue, funcValue), nil
		case "endswith":
			return strings.HasSuffix(strValue, funcValue), nil
		case "contains":
			return strings.Contains(strValue, funcValue), nil
		default:
			return false, fmt.Errorf("unknown function: %s", node.function)
		}
	}

	// Handle comparison operators (only if not a function call)
	if node.operator != "" {
		itemValue, exists := item[node.property]
		if !exists {
			// Property doesn't exist in item
			return false, nil
		}

		return compareValues(itemValue, node.value, node.operator)
	}

	return false, fmt.Errorf("invalid filter node")
}

// getNestedFieldValue gets a field value from a struct by JSON tag name.
// Supports dot-separated nested paths (e.g., "skuId" or "nested.prop").
func getNestedFieldValue(v reflect.Value, jsonPath string) interface{} {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	// Split path for nested access (e.g., "a.b.c")
	parts := strings.Split(jsonPath, ".")
	if len(parts) == 0 {
		return nil
	}

	// Find field by JSON tag name
	targetField := parts[0]
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == targetField {
			fieldValue := v.Field(i)
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					return nil
				}
				fieldValue = fieldValue.Elem()
			}

			if len(parts) > 1 {
				// Recurse for deeper paths
				return getNestedFieldValue(fieldValue, strings.Join(parts[1:], "."))
			}

			if fieldValue.Kind() == reflect.String {
				return fieldValue.String()
			}
			return fieldValue.Interface()
		}
	}

	return nil
}

// compareValues compares two values using the specified operator
func compareValues(a, b interface{}, operator string) (bool, error) {
	// Handle null comparisons
	if a == nil || b == nil {
		switch operator {
		case "eq":
			return a == b, nil
		case "ne":
			return a != b, nil
		default:
			return false, fmt.Errorf("operator %s not supported for null values", operator)
		}
	}

	// Try to compare as strings first
	strA, okA := a.(string)
	strB, okB := b.(string)
	if okA && okB {
		switch operator {
		case "eq":
			return strA == strB, nil
		case "ne":
			return strA != strB, nil
		case "gt":
			return strings.ToLower(strA) > strings.ToLower(strB), nil
		case "ge":
			return strings.ToLower(strA) >= strings.ToLower(strB), nil
		case "lt":
			return strings.ToLower(strA) < strings.ToLower(strB), nil
		case "le":
			return strings.ToLower(strA) <= strings.ToLower(strB), nil
		}
	}

	// Try to compare as booleans
	boolA, okA := a.(bool)
	boolB, okB := b.(bool)
	if okA && okB {
		switch operator {
		case "eq":
			return boolA == boolB, nil
		case "ne":
			return boolA != boolB, nil
		default:
			return false, fmt.Errorf("operator %s not supported for boolean values", operator)
		}
	}

	// Try to compare as numbers
	numA, okA := toFloat64(a)
	numB, okB := toFloat64(b)
	if okA && okB {
		switch operator {
		case "eq":
			return numA == numB, nil
		case "ne":
			return numA != numB, nil
		case "gt":
			return numA > numB, nil
		case "ge":
			return numA >= numB, nil
		case "lt":
			return numA < numB, nil
		case "le":
			return numA <= numB, nil
		}
	}

	return false, fmt.Errorf("cannot compare values of type %T and %T with operator %s", a, b, operator)
}

// toFloat64 converts a value to float64 if possible
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// applySorting sorts items by the specified property or properties
func applySorting(items []map[string]interface{}, orderBy string) {
	if orderBy == "" || len(items) == 0 {
		return
	}

	// Parse order by clause(s) - can be comma-separated
	// Format: "field1 [asc|desc], field2 [asc|desc], ..."
	orderFields := parseOrderByFields(orderBy)
	if len(orderFields) == 0 {
		return
	}

	// Sort the items using a stable sort with a comparator chain
	sort.SliceStable(items, func(i, j int) bool {
		// Compare using all order fields in sequence
		for _, of := range orderFields {
			valI := items[i][of.property]
			valJ := items[j][of.property]

			// Compare this field
			cmp := compareValuesForOrder(valI, valJ, of.ascending)
			if cmp != 0 {
				if of.ascending {
					return cmp < 0
				}
				return cmp > 0
			}
			// Values are equal, continue to next field
		}
		// All fields are equal, maintain original order
		return false
	})
}

// orderField represents a single field in an order by clause
type orderField struct {
	property  string
	ascending bool
}

// parseOrderByFields parses the $orderby parameter into a list of orderField structs
func parseOrderByFields(orderBy string) []orderField {
	result := []orderField{}

	// Split on commas to get individual field specifications
	fields := strings.Split(orderBy, ",")

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		// Split on whitespace to get property and direction
		parts := strings.Fields(field)
		if len(parts) == 0 {
			continue
		}

		property := parts[0]
		ascending := true
		if len(parts) > 1 && strings.ToLower(parts[1]) == "desc" {
			ascending = false
		}

		result = append(result, orderField{
			property:  property,
			ascending: ascending,
		})
	}

	return result
}

// compareValuesForOrder compares two values for sorting purposes
// Returns -1 if a < b, 1 if a > b, 0 if equal
func compareValuesForOrder(a, b interface{}, ascending bool) int {
	// Handle nil values (nulls sort first)
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		if ascending {
			return -1 // nulls first
		}
		return 1 // nulls last
	}
	if b == nil {
		if ascending {
			return 1 // nulls last
		}
		return -1 // nulls first
	}

	// Compare based on type
	switch vi := a.(type) {
	case string:
		if vj, ok := b.(string); ok {
			if vi < vj {
				return -1
			} else if vi > vj {
				return 1
			}
			return 0
		}
	case bool:
		if vj, ok := b.(bool); ok {
			// false before true
			if !vi && vj {
				return -1
			} else if vi && !vj {
				return 1
			}
			return 0
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		numI, okI := toFloat64(a)
		numJ, okJ := toFloat64(b)
		if okI && okJ {
			if numI < numJ {
				return -1
			} else if numI > numJ {
				return 1
			}
			return 0
		}
	}

	// Fallback: convert to string for comparison
	strI := fmt.Sprintf("%v", a)
	strJ := fmt.Sprintf("%v", b)
	if strI < strJ {
		return -1
	} else if strI > strJ {
		return 1
	}
	return 0
}

// applySearch applies OData $search with field-qualified syntax support.
// Supports:
//   - Unqualified: "alice" searches displayName, userPrincipalName, mail, mailNickname
//   - Field-qualified: "displayName:alice" searches only displayName
//   - Multiple terms (space-separated, implicit AND): "displayName:alice mail:saldeti"
//   - Graph API style: "displayName:alice AND userType:Member"
func applySearch(items []map[string]interface{}, search string) []map[string]interface{} {
	if search == "" {
		return items
	}

	// Strip surrounding double-quotes (OData sends $search="..." with the quotes)
	search = strings.Trim(search, `"`)

	// Parse search into individual terms
	terms := parseSearchTerms(search)
	if len(terms) == 0 {
		return items
	}

	// Default fields searched when no field qualifier is given
	defaultFields := []string{"displayName", "userPrincipalName", "mail", "mailNickname"}

	result := make([]map[string]interface{}, 0)
	for _, item := range items {
		allMatch := true
		for _, term := range terms {
			if !matchesSearchTerm(item, term, defaultFields) {
				allMatch = false
				break
			}
		}
		if allMatch {
			result = append(result, item)
		}
	}

	return result
}

// searchTerm represents a parsed $search term with optional field qualifier
type searchTerm struct {
	field string // empty means unqualified (search all default fields)
	value string // the search value (lowercased)
}

// parseSearchTerms parses a search string into individual terms.
// Handles:
//   - "displayName:alice" → {field:"displayName", value:"alice"}
//   - "mail:alice@" → {field:"mail", value:"alice@"}
//   - "alice" → {field:"", value:"alice"}
//   - "displayName:alice AND mail:saldeti" → two terms
//   - Quoted values: "displayName:\"alice bob\"" → {field:"displayName", value:"alice bob"}
func parseSearchTerms(search string) []searchTerm {
	search = strings.TrimSpace(search)
	if search == "" {
		return nil
	}

	var terms []searchTerm

	// Remove explicit AND/OR operators (Graph API uses them but we treat spaces as AND)
	search = strings.ReplaceAll(search, " AND ", " ")
	search = strings.ReplaceAll(search, " and ", " ")
	search = strings.ReplaceAll(search, " OR ", " ")
	search = strings.ReplaceAll(search, " or ", " ")

	// Split by spaces, but respect quoted substrings
	parts := splitSearchParts(search)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for field qualifier (e.g., "displayName:alice")
		if idx := strings.Index(part, ":"); idx > 0 {
			field := part[:idx]
			value := part[idx+1:]
			// Strip surrounding quotes from value
			value = strings.Trim(value, `"`)
			if value != "" {
				terms = append(terms, searchTerm{
					field: field,
					value: strings.ToLower(value),
				})
			}
		} else {
			// Unqualified term
			value := strings.Trim(part, `"`)
			if value != "" {
				terms = append(terms, searchTerm{
					field: "",
					value: strings.ToLower(value),
				})
			}
		}
	}

	return terms
}

// splitSearchParts splits a search string by spaces, respecting quoted substrings
func splitSearchParts(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			inQuotes = !inQuotes
			current.WriteByte(ch)
		} else if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// matchesSearchTerm checks if an item matches a single search term
func matchesSearchTerm(item map[string]interface{}, term searchTerm, defaultFields []string) bool {
	fields := defaultFields
	if term.field != "" {
		fields = []string{term.field}
	}

	for _, field := range fields {
		val, ok := item[field]
		if !ok {
			continue
		}
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(strVal), term.value) {
			return true
		}
	}

	return false
}

// selectFields selects only the specified fields from a map
func selectFields(m map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		if val, ok := m[field]; ok {
			result[field] = val
		}
	}
	return result
}
