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
	"github.com/defenseunicorns/lula/src/pkg/message"
	"github.com/defenseunicorns/lula/src/types"
	"github.com/mitchellh/mapstructure"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
)

// TODO: What is the new version of the information we are displaying on the command line?

func Validate(ctx context.Context, domain string, data map[string]interface{}) (types.Result, error) {
	if domain == "kubernetes" {
		var payload types.Payload
		err := mapstructure.Decode(data, &payload)
		if err != nil {
			return types.Result{}, err
		}

		err = kube.EvaluateWait(payload.Wait)
		if err != nil {
			return types.Result{}, err
		}

		collection, err := kube.QueryCluster(ctx, payload.Resources)
		if err != nil {
			return types.Result{}, err
		}

		// TODO: Add logging optionality for understanding what resources are actually being validated
		results, err := GetValidatedAssets(ctx, payload.Rego, collection)
		if err != nil {
			return types.Result{}, err
		}

		return results, nil

	} else if domain == "api" {
		var payload types.PayloadAPI
		err := mapstructure.Decode(data, &payload)
		if err != nil {
			return types.Result{}, err
		}

		collection := make(map[string]interface{}, 0)

		for _, request := range payload.Requests {
			transport := &http.Transport{}
			client := &http.Client{Transport: transport}

			resp, err := client.Get(request.URL)
			if err != nil {
				return types.Result{}, err
			}
			if resp.StatusCode != 200 {
				return types.Result{},
					fmt.Errorf("expected status code 200 but got %d\n", resp.StatusCode)
			}

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return types.Result{}, err
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType == "application/json" {

				var prettyBuff bytes.Buffer
				json.Indent(&prettyBuff, body, "", "  ")
				prettyJson := prettyBuff.String()

				var tempData interface{}
				err = json.Unmarshal([]byte(prettyJson), &tempData)
				if err != nil {
					return types.Result{}, err
				}
				collection[request.Name] = tempData

			} else {
				return types.Result{}, fmt.Errorf("content type %s is not supported", contentType)
			}
		}

		results, err := GetValidatedAssets(ctx, payload.Rego, collection)
		if err != nil {
			return types.Result{}, err
		}
		return results, nil

	}

	return types.Result{}, fmt.Errorf("domain %s is not supported", domain)
}

// GetValidatedAssets performs the validation of the dataset against the given rego policy
func GetValidatedAssets(ctx context.Context, regoPolicy string, dataset map[string]interface{}) (types.Result, error) {
	message.Debug(message.JSONValue(dataset))
	var matchResult types.Result
	// var trace bool

	buf := topdown.NewBufferTracer()

	// if message.GetLogLevel() == 3 {
	// 	trace = true
	// }

	if len(dataset) == 0 {
		// Not an error but no entries to validate
		message.Warn("no entries to validate")
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
		rego.QueryTracer(buf),
	)

	resultSet, err := regoCalc.Eval(ctx)

	for _, value := range *buf {
		// message.Debug(value.Locals) // contains the payload for the resources but in many different "locals"
		// Ex:	every pod in pods
		// 				podLabel = pod.metadata.labels.foo
		//				podLabel == "bar"
		// { __local3__ = input.podsvt; every __local0__, __local1__ in __local3__ { __local2__ = __local1__.metadata.labels.foo; __local2__ = "bar" } }
		// message.Debug(value.Message) // Nothing too useful
		// message.Debug(message.JSONValue(value.Node)) // Lot of information

		// if value.HasExpr() {
		// 	message.Debug("Has expression")
		// 	if value.HasRule() {
		// 		message.Debug("Has rule")
		// 	}
		// if value.HasBody() {
		// 	message.Debug("Has Body")
		// }
		// }

		// if value.HasBody() {
		// 	message.Debug("Has Body")
		// 	message.Debug(value.LocalMetadata)
		// }

		if value.HasRule() {
			message.Debug("Has rule")
			// message.Debug(value.Ref) // nil

			message.Debug(value.Locals)
			// message.Debug(value.String())
		}

	}

	if err != nil || resultSet == nil || len(resultSet) == 0 {
		return matchResult, fmt.Errorf("failed to evaluate rego policy: %w", err)
	}

	for _, result := range resultSet {
		// message.Debug(result)
		for _, expression := range result.Expressions {
			// message.Debug(expression)
			expressionBytes, err := json.Marshal(expression.Value)
			if err != nil {
				return matchResult, fmt.Errorf("failed to marshal expression: %w", err)
			}

			var expressionMap map[string]interface{}
			err = json.Unmarshal(expressionBytes, &expressionMap)
			if err != nil {
				return matchResult, fmt.Errorf("failed to unmarshal expression: %w", err)
			}
			// check for number of keys in map > 0
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
