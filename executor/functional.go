package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/volcapi/config"
)

var (
	cGreen  = color.New(color.FgHiGreen)
	cRed    = color.New(color.FgHiRed)
	cCyan   = color.New(color.FgHiCyan)
	cYellow = color.New(color.FgHiYellow)
	cGray   = color.New(color.FgHiBlack)
	cBlue   = color.New(color.FgHiBlue)

	symbolPass = cGreen.Sprint("✔")
	symbolFail = cRed.Sprint("✖")
	symbolFile = cYellow.Sprint("📄")
)

func runFunctional(endPoint, method, scenarioName string, scenario config.Scenario) error {
	var totalDuration time.Duration
	label := fmt.Sprintf("%s [%s %s]", scenarioName, method, endPoint)
	spin := showSpinner(label)
	fmt.Printf("%s -%s\n", endPoint, method)
	var bodyBytes []byte
	var err error

	if scenario.Request.Json != nil {
		bodyBytes, err = json.Marshal(scenario.Request.Json)
		if err != nil {
			return err
		}
	} else if scenario.Request.Text != nil {
		bodyBytes = []byte(*scenario.Request.Text)
	} else {
		bodyBytes = []byte("")
	}

	req, err := http.NewRequest(method, endPoint, bytes.NewReader(bodyBytes))
	if err != nil {
		spin.Stop()
		fmt.Printf("   %s %s  %s\n", symbolFail, cRed.Sprintf(scenarioName), cRed.Sprintf("Request build failed"))
		return err
	}
	for key, value := range scenario.Headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	totalDuration += elapsed
	spin.Stop()

	if err != nil {
		fmt.Printf("   %s %s  %s\n", symbolFail, cRed.Sprintf(scenarioName), cRed.Sprintf(err.Error()))
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if scenario.Response.Status != nil && *scenario.Response.Status != resp.StatusCode {
		fmt.Printf("   %s %s  %s (expected status %s, got %s)  %s\n",
			symbolFail,
			cRed.Sprintf(scenarioName),
			cRed.Sprintf("Status mismatch"),
			cGreen.Sprintf("%v", *scenario.Response.Status),
			cRed.Sprintf("%v", resp.StatusCode),
			cGray.Sprintf("(%v)", elapsed.Truncate(time.Millisecond)),
		)
		return nil
	}
	for _, contentType := range resp.Header["Content-Type"] {
		if strings.Contains(contentType, "application/json") {
			if err := validateExpectations(respBody, scenario.Response); err != nil {
				fmt.Printf("   %s %s  %s  %s\n",
					symbolFail,
					cRed.Sprintf(scenarioName),
					cRed.Sprintf(err.Error()),
					cGray.Sprintf("(%v)", elapsed.Truncate(time.Millisecond)),
				)
				return nil
			}
		}
		if scenario.Response.Body.Text != nil {
			bodyStr := string(respBody)
			txt := *scenario.Response.Body.Text
			if !strings.Contains(bodyStr, txt) {
				return fmt.Errorf("expected %q in response body, but not found", txt)
			}
		}
	}

	fmt.Printf("   %s %s  %s\n",
		symbolPass,
		cGreen.Sprintf(scenarioName),
		cGray.Sprintf("(%v)", elapsed.Truncate(time.Millisecond)))
	return nil
}

func validateExpectations(respBody []byte, expect config.Response) error {
	var actualBody map[string]any
	if err := json.Unmarshal(respBody, &actualBody); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}
	for _, path := range expect.Body.Contains {
		val, ok := getByPath(actualBody, path)
		if !ok {
			return fmt.Errorf("expected %s to exist, but it does NOT", path)
		}
		if val == nil {
			return fmt.Errorf("expected %s to exist, but it is NULL", path)
		}
	}

	for key, value := range actualBody {
		if expect.Body.Json != nil {
			json := *expect.Body.Json
			if expected, ok := json[key]; ok {
				if err := validateJNode(value, key, expected); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateJNode(actual any, key string, expected config.JNode) error {
	switch realValue := actual.(type) {
	case string:
		expectedValue, match := expected.Value.(string)

		if expected.Type != nil && *expected.Type != "string" || expectedValue != "" && !match {
			return fmt.Errorf("invalid type expected %s to be string", key)
		}
		if expectedValue != "" && expectedValue != realValue {
			return fmt.Errorf("expected %s to be '%s', got '%s'", key, expectedValue, realValue)
		}
		if expected.Min != nil && len(realValue) < *expected.Min {
			return fmt.Errorf("expected %s to have string less than %v", key, expected.Min)
		}
		if expected.Max != nil && len(realValue) > *expected.Max {
			return fmt.Errorf("expected %s to have string bigger than %v", key, expected.Max)
		}
	case float64:
		if exp, ok := expected.Value.(float64); ok && exp != realValue {
			return fmt.Errorf("expected %s to be %v, got %v", key, exp, realValue)
		}
	case bool:
		if expected.Value != nil {
			if exp, ok := expected.Value.(bool); ok && exp != realValue {
				return fmt.Errorf("expected %s to be %v, got %v", key, exp, realValue)
			}
		}
	case map[string]any:
		return validateObject(realValue, expected.Object)
	case []any:
		return validateList(realValue, expected.List)
	default:
		return fmt.Errorf("unsupported type: %T", realValue)

	}
	return nil
}

func validateObject(actual map[string]any, expected map[string]config.JNode) error {
	for k, expNode := range expected {
		actVal, ok := actual[k]
		if !ok {
			return fmt.Errorf("missing key %s", k)
		}
		// Recurse using same validation rules
		if err := validateJNode(actVal, k, expNode); err != nil {
			return fmt.Errorf("%s: %v", k, err)
		}
	}
	return nil
}

func validateList(actual []any, expected []map[string]config.JNode) error {
	// For now: simple length check
	if len(expected) > 0 && len(actual) != len(expected) {
		return fmt.Errorf("array length mismatch (expected %d, got %d)", len(expected), len(actual))
	}
	// If expected has schema for items → validate
	for i, act := range actual {
		if len(expected) > 0 {
			if err := validateObject(asMap(act), expected[i]); err != nil {
				return fmt.Errorf("index %d: %v", i, err)
			}
		}
	}
	return nil
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

type pathToken struct {
	key   string
	index *int
}

var (
	reBracket = regexp.MustCompile(`^([a-zA-Z0-9_-]+)\[(\d+)\]$`)
)

func getByPath(data any, path string) (any, bool) {
	tokens := parsePath(path)
	current := data

	for _, tok := range tokens {
		switch node := current.(type) {
		case map[string]any:
			val, ok := node[tok.key]
			if !ok {
				return nil, false
			}
			current = val

		case []any:
			if tok.index == nil || *tok.index >= len(node) {
				return nil, false
			}
			current = node[*tok.index]

		default:
			return nil, false
		}
	}

	return current, true
}

func parsePath(path string) []pathToken {
	parts := strings.Split(path, ".")
	var tokens []pathToken
	for _, p := range parts {
		if m := reBracket.FindStringSubmatch(p); m != nil {
			idx, _ := strconv.Atoi(m[2])
			tokens = append(tokens, pathToken{key: m[1], index: &idx})
		} else {
			tokens = append(tokens, pathToken{key: p})
		}
	}
	return tokens
}

func showSpinner(label string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + cCyan.Sprintf(label)
	s.Color("cyan")
	s.Start()
	return s
}
