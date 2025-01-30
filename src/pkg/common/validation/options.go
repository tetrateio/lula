package validation

import (
	"fmt"

	"github.com/defenseunicorns/lula/src/pkg/common/composition"
	"github.com/defenseunicorns/lula/src/pkg/message"
)

type Option func(*Validator) error

func WithComposition(composer *composition.Composer, path string) Option {
	return func(v *Validator) error {
		var err error
		if composer == nil {
			composer, err = composition.New(composition.WithModelFromLocalPath(path))
			if err != nil {
				return fmt.Errorf("error creating composition context: %v", err)
			}
		}
		v.composer = composer
		return nil
	}
}

func WithAllowExecution(confirmExecution, runNonInteractively bool) Option {
	return func(v *Validator) error {
		if !confirmExecution {
			if !runNonInteractively {
				v.requestExecutionConfirmation = true
			} else {
				message.Infof("Validations requiring execution will NOT be run")
			}
		} else {
			v.runExecutableValidations = true
		}
		return nil
	}
}

func WithOutputDir(rootDir string) Option {
	return func(v *Validator) error {
		v.outputsDir = rootDir
		return nil
	}
}

func WithSaveResources(saveResources bool) Option {
	return func(v *Validator) error {
		v.saveResources = saveResources
		return nil
	}
}

func WithTests(runTests bool) Option {
	return func(v *Validator) error {
		v.runTests = runTests
		return nil
	}
}
