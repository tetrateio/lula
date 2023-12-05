package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	kube "github.com/defenseunicorns/lula/src/pkg/common/kubernetes"
	"github.com/defenseunicorns/lula/src/types"
	"github.com/mitchellh/mapstructure"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

// TODO: What is the new version of the information we are displaying on the command line?

func InterrogateKubernetes(
	ctx context.Context,
	data map[string]interface{},
) (
	context.Context,
	string,
	map[string]interface{},
	error,
) {
		var payload types.Payload
		err := mapstructure.Decode(data, &payload)
		if err != nil {
			return nil, "", nil, err
		}
		collection, err := kube.QueryCluster(ctx, payload.Resources)
		if err != nil {
			return nil, "", nil, err
		}

		return ctx, payload.Rego, collection, nil
}

func InterrogateAPI(
	ctx context.Context,
	data map[string]interface{},
) (
	context.Context,
	string,
	map[string]interface{},
	error,
) {
		var payload types.PayloadAPI
		err := mapstructure.Decode(data, &payload)
		if err != nil {
			return nil, "", nil, err
		}

		collection := make(map[string]interface{}, 0)

		for _, request := range payload.Requests {
			transport := &http.Transport{}
			client := &http.Client{Transport: transport}

			resp, err := client.Get(request.URL)
			if err != nil {
				return nil, "", nil, err
			}
			if resp.StatusCode != 200 {
				return nil, "", nil, fmt.Errorf("expected status code 200 but got %d\n", resp.StatusCode)
			}

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, "", nil, err
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType == "application/json" {

				var prettyBuff bytes.Buffer
				json.Indent(&prettyBuff, body, "", "  ")
				prettyJson := prettyBuff.String()

				var tempData interface{}
				err = json.Unmarshal([]byte(prettyJson), &tempData)
				if err != nil {
					return nil, "", nil, err
				}
				collection[request.Name] = tempData

			} else {
				return nil, "", nil, fmt.Errorf("content type %s is not supported", contentType)
			}
		}

	return ctx, payload.Rego, collection, nil
}

func Validate(ctx context.Context, domain string, data map[string]interface{}) (types.Result, error) {
	var rego string
	collection := make(map[string]interface{}, 0)
	var err error

	switch domain {
	case "kubernetes":
		ctx, rego, collection, err = InterrogateKubernetes(ctx, data)
		if err != nil {
			return types.Result{}, err
		}

	case "api":
		ctx, rego, collection, err = InterrogateAPI(ctx, data)
		if err != nil {
			return types.Result{}, err
		}

	default:
		return types.Result{}, fmt.Errorf("domain %s is not supported", domain)
	}

	//TODO: Add logging optionality for understanding what resources are actually being validated
	results, err := GetValidatedAssets(ctx, rego, collection)
	if err != nil {
		return types.Result{}, err
	}
	return results, nil
}

// GetValidatedAssets performs the validation of the dataset against the given rego policy
func GetValidatedAssets(ctx context.Context, regoPolicy string, dataset map[string]interface{}) (types.Result, error) {
	var matchResult types.Result

	if len(dataset) == 0 {
		// Not an error but no entries to validate
		// TODO: add a warning log
		return matchResult, nil
	}

	compiler, err := ast.CompileModules(map[string]string{
		"validate.rego": regoPolicy,
	})
	if err != nil {
		log.Fatal(err)
		return matchResult, fmt.Errorf("failed to compile rego policy: %w", err)
	}

	regoCalc := rego.New(
		rego.Query("data.validate"),
		rego.Compiler(compiler),
		rego.Input(dataset),
	)

	resultSet, err := regoCalc.Eval(ctx)

	if err != nil || resultSet == nil || len(resultSet) == 0 {
		return matchResult, fmt.Errorf("failed to evaluate rego policy: %w", err)
	}

	for _, result := range resultSet {
		for _, expression := range result.Expressions {
			expressionBytes, err := json.Marshal(expression.Value)
			if err != nil {
				return matchResult, fmt.Errorf("failed to marshal expression: %w", err)
			}

			var expressionMap map[string]interface{}
			err = json.Unmarshal(expressionBytes, &expressionMap)
			if err != nil {
				return matchResult, fmt.Errorf("failed to unmarshal expression: %w", err)
			}
			// TODO: add logging optionality here for developer experience
			if matched, ok := expressionMap["validate"]; ok && matched.(bool) {
				// TODO: Is there a way to determine how many resources failed?
				matchResult.Passing += 1
			} else {
				matchResult.Failing += 1
			}
		}
	}

	return matchResult, nil
}
