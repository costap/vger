package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// flexibleStringSlice is a slice of strings that can unmarshal from:
// - a JSON string: "foo" -> []string{"foo"}
// - a JSON array of strings: ["foo", "bar"] -> []string{"foo", "bar"}
// - a JSON array of objects: [{"name": "foo", "description": "bar"}] -> []string{"foo — bar"}
// - a mixed JSON array: ["foo", {"name": "bar"}] -> []string{"foo", "bar"}
// - a single JSON object: {"name": "foo"} -> []string{"foo"}
// - null or empty -> nil
type flexibleStringSlice []string

var primaryKeyCandidates = []string{
	"name",
	"approach",
	"technology",
	"tool",
	"project",
	"title",
	"speaker",
	"step",
	"action",
	"item",
}

var secondaryKeyCandidates = []string{
	"description",
	"desc",
	"tradeoffs",
	"trade_offs",
	"details",
	"summary",
	"relevance",
	"why",
	"reason",
	"notes",
	"rationale",
	"affiliation",
}

func (f *flexibleStringSlice) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*f = nil
		return nil
	}

	// 1. Single string: "..."
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s != "" {
			*f = []string{s}
		} else {
			*f = nil
		}
		return nil
	}

	// 2. Single object: {...}
	if data[0] == '{' {
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		if formatted := formatObjectAsString(obj); formatted != "" {
			*f = []string{formatted}
		} else {
			*f = nil
		}
		return nil
	}

	// 3. Array: [...]
	if data[0] == '[' {
		var rawList []json.RawMessage
		if err := json.Unmarshal(data, &rawList); err != nil {
			return err
		}
		result := make([]string, 0, len(rawList))
		for _, raw := range rawList {
			raw = bytes.TrimSpace(raw)
			if len(raw) == 0 || string(raw) == "null" {
				continue
			}
			if raw[0] == '"' {
				var s string
				if err := json.Unmarshal(raw, &s); err == nil {
					s = strings.TrimSpace(s)
					if s != "" {
						result = append(result, s)
					}
				}
			} else if raw[0] == '{' {
				var obj map[string]any
				if err := json.Unmarshal(raw, &obj); err == nil {
					if formatted := formatObjectAsString(obj); formatted != "" {
						result = append(result, formatted)
					}
				}
			} else {
				var val any
				if err := json.Unmarshal(raw, &val); err == nil {
					if formatted := formatValue(val); formatted != "" {
						result = append(result, formatted)
					}
				}
			}
		}
		*f = result
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into flexible string slice", string(data))
}

func findKey(obj map[string]any, candidates []string) (string, any, bool) {
	for _, c := range candidates {
		for k, v := range obj {
			if strings.EqualFold(k, c) {
				return k, v, true
			}
		}
	}
	return "", nil, false
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case []any:
		strs := make([]string, 0, len(val))
		for _, item := range val {
			if s := formatValue(item); s != "" {
				strs = append(strs, s)
			}
		}
		return strings.Join(strs, ", ")
	case map[string]any:
		return formatObjectAsString(val)
	default:
		return strings.TrimSpace(fmt.Sprint(val))
	}
}

func formatObjectAsString(obj map[string]any) string {
	if len(obj) == 0 {
		return ""
	}

	primKey, primVal, hasPrim := findKey(obj, primaryKeyCandidates)
	secKey, secVal, hasSec := findKey(obj, secondaryKeyCandidates)

	if hasPrim && hasSec {
		primStr := formatValue(primVal)
		secStr := formatValue(secVal)

		// Special case: speaker + affiliation -> "Jane Smith (Red Hat)"
		if strings.EqualFold(primKey, "speaker") && strings.EqualFold(secKey, "affiliation") {
			if secStr != "" {
				return fmt.Sprintf("%s (%s)", primStr, secStr)
			}
			return primStr
		}

		if primStr != "" && secStr != "" {
			return fmt.Sprintf("%s — %s", primStr, secStr)
		}
		if primStr != "" {
			return primStr
		}
		return secStr
	}

	if hasPrim {
		primStr := formatValue(primVal)
		if len(obj) == 1 || primStr == "" {
			return primStr
		}
		// Include remaining fields
		keys := make([]string, 0, len(obj))
		for k := range obj {
			if !strings.EqualFold(k, primKey) {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(obj[k])))
		}
		return fmt.Sprintf("%s — %s", primStr, strings.Join(parts, ", "))
	}

	if hasSec && len(obj) == 1 {
		return formatValue(secVal)
	}

	// Single key-value pair without known candidate names
	if len(obj) == 1 {
		for k, v := range obj {
			valStr := formatValue(v)
			if valStr != "" {
				return fmt.Sprintf("%s — %s", k, valStr)
			}
			return k
		}
	}

	// Multiple arbitrary fields: sort keys and join
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(obj[k])))
	}
	return strings.Join(parts, ", ")
}
