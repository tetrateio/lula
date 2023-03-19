package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/defenseunicorns/lula/src/internal/types"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

func GetMatchedAssets(ctx context.Context, regoPolicy string, dataset []map[string]interface{}) (matchResult types.Results, err error) {
	var wg sync.WaitGroup
	compiler, err := ast.CompileModules(map[string]string{
		"match.rego": regoPolicy,
	})
	if err != nil {
		log.Fatal(err)
		return matchResult, fmt.Errorf("failed to compile rego policy: %w", err)
	}

	for _, asset := range dataset {
		wg.Add(1)
		go func(asset map[string]interface{}) {
			defer wg.Done()

			regoCalc := rego.New(
				rego.Query("data.match"),
				rego.Compiler(compiler),
				rego.Input(asset),
			)

			resultSet, err := regoCalc.Eval(ctx)
			fmt.Println()
			if err != nil || resultSet == nil || len(resultSet) == 0 {
				wg.Done()
			}

			for _, result := range resultSet {
				fmt.Println()
				for _, expression := range result.Expressions {
					expressionBytes, err := json.Marshal(expression.Value)
					if err != nil {
						wg.Done()
					}

					var expressionMap map[string]interface{}
					err = json.Unmarshal(expressionBytes, &expressionMap)
					if err != nil {
						wg.Done()
					}

					if matched, ok := expressionMap["match"]; ok && matched.(bool) {
						fmt.Printf("Asset matched policy: %s/%s\n\n", expression, asset)
						matchResult.Match += 1
					} else {
						fmt.Printf("Asset no matched policy: %s/%s\n\n", expression, asset)
						matchResult.NonMatch += 1
					}
				}
			}
		}(asset)
	}

	wg.Wait()

	return matchResult, nil
}
