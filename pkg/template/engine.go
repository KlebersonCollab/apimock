package template

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RequestContext holds context of the incoming HTTP request for template interpolation
type RequestContext struct {
	Params      map[string]string   `json:"params"`
	Query       map[string][]string `json:"query"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body"`
	ParsedBody  interface{}         `json:"parsedBody,omitempty"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	ClientIP    string              `json:"clientIp"`
	Timestamp   string              `json:"timestamp"`
}

// Engine processes templates with faker data and request context
type Engine struct {
	faker *Faker
}

// NewEngine creates a new template engine instance
func NewEngine() *Engine {
	return &Engine{
		faker: NewFaker(),
	}
}

var (
	tagRegex    = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_\.\(\)\,\'\-\@\s]+)\s*\}\}`)
	repeatRegex = regexp.MustCompile(`(?s)\{\{#repeat\s+([0-9]+(?:\s*,\s*[0-9]+)?)\s*\}\}(.*?)\{\{/repeat\}\}`)
)

// Evaluate parses and replaces all template tags in the input string
func (e *Engine) Evaluate(input string, ctx *RequestContext) string {
	if strings.TrimSpace(input) == "" {
		return input
	}

	if ctx == nil {
		ctx = &RequestContext{
			Params:    make(map[string]string),
			Query:     make(map[string][]string),
			Headers:   make(map[string][]string),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// 1. Process {{#repeat N}} ... {{/repeat}} blocks
	result := e.processRepeatBlocks(input, ctx)

	// 2. Process all dynamic interpolation tags {{ ... }}
	result = tagRegex.ReplaceAllStringFunc(result, func(match string) string {
		sub := tagRegex.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		rawExpr := strings.TrimSpace(sub[1])
		val, ok := e.resolveExpression(rawExpr, ctx, 0, 1)
		if ok {
			return val
		}
		return match
	})

	return result
}

// processRepeatBlocks expands array repeat helpers
func (e *Engine) processRepeatBlocks(input string, ctx *RequestContext) string {
	return repeatRegex.ReplaceAllStringFunc(input, func(match string) string {
		sub := repeatRegex.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		rangeSpec := strings.TrimSpace(sub[1])
		innerTpl := sub[2]

		count := 3
		if strings.Contains(rangeSpec, ",") {
			parts := strings.Split(rangeSpec, ",")
			min, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			max, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			count = e.faker.Number(min, max)
		} else {
			if n, err := strconv.Atoi(rangeSpec); err == nil && n >= 0 {
				count = n
			}
		}

		var items []string
		for i := 0; i < count; i++ {
			iterationTpl := innerTpl
			// Replace {{@index}} and {{@iteration}}
			iterationTpl = strings.ReplaceAll(iterationTpl, "{{@index}}", strconv.Itoa(i))
			iterationTpl = strings.ReplaceAll(iterationTpl, "{{@iteration}}", strconv.Itoa(i+1))

			// Resolve tags inside iteration
			expanded := tagRegex.ReplaceAllStringFunc(iterationTpl, func(tagMatch string) string {
				subTag := tagRegex.FindStringSubmatch(tagMatch)
				if len(subTag) < 2 {
					return tagMatch
				}
				rawExpr := strings.TrimSpace(subTag[1])
				if rawExpr == "@index" {
					return strconv.Itoa(i)
				}
				if rawExpr == "@iteration" {
					return strconv.Itoa(i + 1)
				}
				val, ok := e.resolveExpression(rawExpr, ctx, i, i+1)
				if ok {
					return val
				}
				return tagMatch
			})
			items = append(items, strings.TrimSpace(expanded))
		}

		return strings.Join(items, ",\n    ")
	})
}

// resolveExpression evaluates a single tag expression
func (e *Engine) resolveExpression(expr string, ctx *RequestContext, index, iteration int) (string, bool) {
	expr = strings.TrimSpace(expr)

	// Special loop tokens
	if expr == "@index" {
		return strconv.Itoa(index), true
	}
	if expr == "@iteration" {
		return strconv.Itoa(iteration), true
	}

	// 1. Request context expressions: req.*
	if strings.HasPrefix(expr, "req.") {
		return e.resolveReqExpression(expr[4:], ctx)
	}

	// 2. Faker expressions: faker.*
	if strings.HasPrefix(expr, "faker.") {
		return e.resolveFakerExpression(expr[6:])
	}

	return "", false
}

// resolveReqExpression resolves request parameters, query, body, and headers
func (e *Engine) resolveReqExpression(expr string, ctx *RequestContext) (string, bool) {
	if ctx == nil {
		return "", false
	}

	switch expr {
	case "method":
		return ctx.Method, true
	case "path":
		return ctx.Path, true
	case "clientIp", "clientIP":
		return ctx.ClientIP, true
	case "timestamp":
		if ctx.Timestamp != "" {
			return ctx.Timestamp, true
		}
		return time.Now().Format(time.RFC3339), true
	}

	// req.params.<param_name>
	if strings.HasPrefix(expr, "params.") {
		key := expr[7:]
		if val, exists := ctx.Params[key]; exists {
			return val, true
		}
		return "", true
	}

	// req.query.<query_name>
	if strings.HasPrefix(expr, "query.") {
		key := expr[6:]
		if vals, exists := ctx.Query[key]; exists && len(vals) > 0 {
			return vals[0], true
		}
		return "", true
	}

	// req.headers.<header_name>
	if strings.HasPrefix(expr, "headers.") {
		key := strings.ToLower(expr[8:])
		for hKey, vals := range ctx.Headers {
			if strings.ToLower(hKey) == key && len(vals) > 0 {
				return vals[0], true
			}
		}
		return "", true
	}

	// req.body.<field_path>
	if strings.HasPrefix(expr, "body.") {
		fieldPath := expr[5:]
		if ctx.ParsedBody == nil && ctx.Body != "" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(ctx.Body), &parsed); err == nil {
				ctx.ParsedBody = parsed
			}
		}
		if ctx.ParsedBody != nil {
			val := extractJSONPath(ctx.ParsedBody, fieldPath)
			if val != nil {
				switch v := val.(type) {
				case string:
					return v, true
				case float64:
					if v == float64(int64(v)) {
						return strconv.FormatInt(int64(v), 10), true
					}
					return strconv.FormatFloat(v, 'f', -1, 64), true
				case bool:
					return strconv.FormatBool(v), true
				default:
					b, _ := json.Marshal(v)
					return string(b), true
				}
			}
		}
		return "", true
	}

	return "", false
}

// resolveFakerExpression resolves faker methods with optional parameters
func (e *Engine) resolveFakerExpression(expr string) (string, bool) {
	// Parse function calls with arguments e.g. number(10, 50), price(5.0, 20.0), lorem(10)
	fnName := expr
	var args []string

	if idx := strings.Index(expr, "("); idx != -1 && strings.HasSuffix(expr, ")") {
		fnName = expr[:idx]
		argsStr := expr[idx+1 : len(expr)-1]
		if strings.TrimSpace(argsStr) != "" {
			rawArgs := strings.Split(argsStr, ",")
			for _, arg := range rawArgs {
				cleanArg := strings.TrimSpace(arg)
				cleanArg = strings.Trim(cleanArg, `"'`)
				args = append(args, cleanArg)
			}
		}
	}

	switch strings.ToLower(fnName) {
	case "name":
		return e.faker.Name(), true
	case "firstname", "first_name":
		return e.faker.FirstName(), true
	case "lastname", "last_name":
		return e.faker.LastName(), true
	case "username":
		return e.faker.Username(), true
	case "email":
		return e.faker.Email(), true
	case "avatar":
		return e.faker.Avatar(), true
	case "phone":
		return e.faker.Phone(), true
	case "jobtitle", "job_title":
		return e.faker.JobTitle(), true
	case "company":
		return e.faker.Company(), true
	case "street":
		return e.faker.Street(), true
	case "city":
		return e.faker.City(), true
	case "country":
		return e.faker.Country(), true
	case "zipcode", "zip_code":
		return e.faker.ZipCode(), true
	case "uuid":
		return e.faker.UUID(), true
	case "id":
		return strconv.Itoa(e.faker.ID()), true
	case "boolean", "bool":
		return strconv.FormatBool(e.faker.Boolean()), true
	case "status":
		return e.faker.Status(), true
	case "role":
		return e.faker.Role(), true
	case "currency":
		return e.faker.Currency(), true
	case "date":
		return e.faker.Date(), true
	case "datetime", "isodate":
		return e.faker.DateTime(), true
	case "sentence":
		return e.faker.Sentence(), true
	case "paragraph":
		return e.faker.Paragraph(), true
	case "lorem":
		count := 5
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil {
				count = n
			}
		}
		return e.faker.Lorem(count), true
	case "number", "int":
		min, max := 1, 100
		if len(args) >= 2 {
			min, _ = strconv.Atoi(args[0])
			max, _ = strconv.Atoi(args[1])
		} else if len(args) == 1 {
			max, _ = strconv.Atoi(args[0])
			min = 1
		}
		return strconv.Itoa(e.faker.Number(min, max)), true
	case "price", "float":
		min, max := 10.0, 500.0
		if len(args) >= 2 {
			min, _ = strconv.ParseFloat(args[0], 64)
			max, _ = strconv.ParseFloat(args[1], 64)
		}
		return fmt.Sprintf("%.2f", e.faker.Price(min, max)), true
	case "pastdate", "past_date":
		days := 30
		if len(args) > 0 {
			days, _ = strconv.Atoi(args[0])
		}
		return e.faker.PastDate(days), true
	case "futuredate", "future_date":
		days := 30
		if len(args) > 0 {
			days, _ = strconv.Atoi(args[0])
		}
		return e.faker.FutureDate(days), true
	case "image":
		w, h := 600, 400
		cat := ""
		if len(args) >= 2 {
			w, _ = strconv.Atoi(args[0])
			h, _ = strconv.Atoi(args[1])
		}
		if len(args) >= 3 {
			cat = args[2]
		}
		return e.faker.Image(w, h, cat), true
	}

	return "", false
}

// extractJSONPath traverses nested maps or slices given a dot-delimited path (e.g., "user.address.city")
func extractJSONPath(data interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	curr := data

	for _, part := range parts {
		if curr == nil {
			return nil
		}
		switch m := curr.(type) {
		case map[string]interface{}:
			var ok bool
			curr, ok = m[part]
			if !ok {
				return nil
			}
		default:
			return nil
		}
	}
	return curr
}
