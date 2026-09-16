package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"apimock/pkg/models"
)

// EvaluateScenarios checks each configured scenario in priority order and returns the first matching one
func EvaluateScenarios(r *http.Request, ep *models.Endpoint, params map[string]string, reqBody string) *models.Scenario {
	if ep == nil || len(ep.Scenarios) == 0 {
		return nil
	}

	var parsedBody interface{}
	var bodyParsed bool

	for i := range ep.Scenarios {
		sc := &ep.Scenarios[i]
		if !sc.Enabled {
			continue
		}

		if len(sc.Conditions) == 0 {
			// A scenario with no conditions never automatically triggers
			continue
		}

		// Lazy parse JSON body on first condition that requires body inspection
		if !bodyParsed {
			needsBody := false
			for _, c := range sc.Conditions {
				if c.Source == models.ConditionSourceBody {
					needsBody = true
					break
				}
			}
			if needsBody && strings.TrimSpace(reqBody) != "" {
				_ = json.Unmarshal([]byte(reqBody), &parsedBody)
				bodyParsed = true
			}
		}

		matched := evaluateScenarioMatch(r, sc, params, parsedBody)
		if matched {
			return sc
		}
	}

	return nil
}

func evaluateScenarioMatch(r *http.Request, sc *models.Scenario, params map[string]string, parsedBody interface{}) bool {
	mode := sc.MatchMode
	if mode == "" {
		mode = models.MatchModeAll
	}
	mode = strings.ToLower(mode)

	for _, cond := range sc.Conditions {
		actual, exists := extractActualValue(r, cond, params, parsedBody)
		condPassed := evaluateCondition(actual, exists, cond)

		if mode == models.MatchModeAny {
			if condPassed {
				return true
			}
		} else {
			// MatchModeAll
			if !condPassed {
				return false
			}
		}
	}

	// For MatchModeAll, if we got here and checked conditions, it's true
	// For MatchModeAny, if we got here and none passed, it's false
	return mode == models.MatchModeAll
}

func extractActualValue(r *http.Request, cond models.Condition, params map[string]string, parsedBody interface{}) (string, bool) {
	prop := strings.TrimSpace(cond.Property)

	switch cond.Source {
	case models.ConditionSourceQuery:
		query := r.URL.Query()
		if vals, ok := query[prop]; ok && len(vals) > 0 {
			return vals[0], true
		}
		return "", false

	case models.ConditionSourceHeader:
		propLower := strings.ToLower(prop)
		for hKey, hVals := range r.Header {
			if strings.ToLower(hKey) == propLower && len(hVals) > 0 {
				return hVals[0], true
			}
		}
		return "", false

	case models.ConditionSourceParam:
		if val, ok := params[prop]; ok {
			return val, true
		}
		return "", false

	case models.ConditionSourceBody:
		if parsedBody == nil {
			return "", false
		}
		val, exists := extractJSONProperty(parsedBody, prop)
		if !exists {
			return "", false
		}
		return formatVal(val), true
	}

	return "", false
}

func evaluateCondition(actual string, exists bool, cond models.Condition) bool {
	target := cond.Value

	switch cond.Operator {
	case models.OperatorEquals:
		return exists && actual == target

	case models.OperatorNotEquals:
		return !exists || actual != target

	case models.OperatorContains:
		return exists && strings.Contains(strings.ToLower(actual), strings.ToLower(target))

	case models.OperatorRegex:
		if !exists {
			return false
		}
		re, err := regexp.Compile(target)
		if err != nil {
			return false
		}
		return re.MatchString(actual)

	case models.OperatorGt:
		if !exists {
			return false
		}
		fActual, err1 := strconv.ParseFloat(actual, 64)
		fTarget, err2 := strconv.ParseFloat(target, 64)
		if err1 == nil && err2 == nil {
			return fActual > fTarget
		}
		return actual > target

	case models.OperatorGte:
		if !exists {
			return false
		}
		fActual, err1 := strconv.ParseFloat(actual, 64)
		fTarget, err2 := strconv.ParseFloat(target, 64)
		if err1 == nil && err2 == nil {
			return fActual >= fTarget
		}
		return actual >= target

	case models.OperatorLt:
		if !exists {
			return false
		}
		fActual, err1 := strconv.ParseFloat(actual, 64)
		fTarget, err2 := strconv.ParseFloat(target, 64)
		if err1 == nil && err2 == nil {
			return fActual < fTarget
		}
		return actual < target

	case models.OperatorLte:
		if !exists {
			return false
		}
		fActual, err1 := strconv.ParseFloat(actual, 64)
		fTarget, err2 := strconv.ParseFloat(target, 64)
		if err1 == nil && err2 == nil {
			return fActual <= fTarget
		}
		return actual <= target

	case models.OperatorIsEmpty:
		return !exists || strings.TrimSpace(actual) == ""

	case models.OperatorIsNotEmpty:
		return exists && strings.TrimSpace(actual) != ""
	}

	return false
}

func extractJSONProperty(data interface{}, path string) (interface{}, bool) {
	if data == nil || path == "" {
		return nil, false
	}

	parts := strings.Split(path, ".")
	curr := data

	for _, part := range parts {
		if curr == nil {
			return nil, false
		}
		switch m := curr.(type) {
		case map[string]interface{}:
			val, ok := m[part]
			if !ok {
				return nil, false
			}
			curr = val
		default:
			return nil, false
		}
	}

	return curr, true
}

func formatVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
