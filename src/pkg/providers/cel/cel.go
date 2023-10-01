package cel

import (
	"fmt"

	// "github.com/accuknox/kubernetes-cel-validator/resource-cel-validator/pkg"
	"github.com/accuknox/kubernetes-cel-validator/resource-cel-validator/pkg/types"
	"github.com/mitchellh/mapstructure"
)

func Validate(data map[string]interface{}) error {

	var precondition types.KubernetesResourcePrecondition
	err := mapstructure.Decode(data, &precondition)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", precondition)

	// prep for calling this cel function
	// pkg.GetKubernetesResourcePreconditionResult()

	return nil
}
