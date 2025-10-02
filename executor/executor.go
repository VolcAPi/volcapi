package executor

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/volcapi/config"
)

func Run(config *config.Config) error {
	for api, detail := range config.Path {
		endPoint := config.Host + api
		for _, scenarioName := range detail.Scenarios {
			if scenario, ok := config.Scenarios[scenarioName]; ok {
				var bufEndpoin bytes.Buffer
				path := replacePathParams(endPoint, scenario.Params)
				bufEndpoin.WriteString(path)
				for key, value := range scenario.Query {
					bufEndpoin.WriteString(fmt.Sprintf("?%s=%s", key, value))
				}
				if err := runFunctional(bufEndpoin.String(), detail.Method, scenarioName, scenario); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func replacePathParams(path string, params map[string]string) string {
	// regex to find {param} patterns
	re := regexp.MustCompile(`\{([^}]+)\}`)

	// replace each placeholder
	result := re.ReplaceAllStringFunc(path, func(match string) string {
		// extract param name (remove { and })
		key := strings.Trim(match, "{}")
		if val, ok := params[key]; ok {
			return val
		}
		return "" // no match, replace with empty string
	})

	return result
}
