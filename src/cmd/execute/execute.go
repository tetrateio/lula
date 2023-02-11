package execute

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	types "github.com/defenseunicorns/lula/src/internal/types"
	yaml2 "github.com/ghodss/yaml"
	sanitizederror "github.com/kyverno/kyverno/cmd/cli/kubectl-kyverno/utils/sanitizedError"
	"github.com/kyverno/kyverno/cmd/cli/kubectl-kyverno/utils/store"
	"github.com/kyverno/kyverno/pkg/clients/dclient"
	"github.com/kyverno/kyverno/pkg/config"
	"github.com/kyverno/kyverno/pkg/openapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	log "sigs.k8s.io/controller-runtime/pkg/log"
	yaml1 "sigs.k8s.io/yaml"

	// Kyverno Dependencies below
	"github.com/go-git/go-billy/v5/memfs"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	v1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"github.com/kyverno/kyverno/api/kyverno/v1beta1"
	"github.com/kyverno/kyverno/cmd/cli/kubectl-kyverno/utils/common"
	policy2 "github.com/kyverno/kyverno/pkg/policy"
	gitutils "github.com/kyverno/kyverno/pkg/utils/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Resource struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

type Policy struct {
	Name      string     `json:"name"`
	Resources []Resource `json:"resources"`
}

type Values struct {
	Policies []Policy `json:"policies"`
}

type SkippedInvalidPolicies struct {
	skipped []string
	invalid []string
}

type ApplyCommandConfig struct {
	KubeConfig      string
	Context         string
	Namespace       string
	MutateLogPath   string
	VariablesString string
	ValuesFile      string
	UserInfoPath    string
	Cluster         bool
	PolicyReport    bool
	Stdin           bool
	RegistryAccess  bool
	AuditWarn       bool
	ResourcePaths   []string
	PolicyPaths     []string
	GitBranch       string
	warnExitCode    int
}

var executeHelp = `
To execute on a cluster:
	lula execute ./oscal-component.yaml

To execute on a resource:
	lula execute ./oscal-component.yaml -r resource.yaml

To execute without creation of any report files
	lula execute ./oscal-component.yaml -d
`

var generateHelp = `
To generate kyverno policies:
	lula generate ./oscal-component.yaml -o ./out
`

var resourcePaths []string
var cluster, dryRun bool

var osExit = os.Exit

var executeCmd = &cobra.Command{
	Use:     "execute",
	Short:   "execute",
	Example: executeHelp,
	Run: func(cmd *cobra.Command, componentDefinitionPaths []string) {
		// Conduct further error checking here (IE flags/arguments)
		if len(componentDefinitionPaths) == 0 {
			fmt.Println("Path to the local OSCAL file must be present")
			os.Exit(1)
		}

		err := conductExecute(componentDefinitionPaths, resourcePaths, dryRun)
		if err != nil {
			log.Log.Error(err, "error string")
		}
	},
}

func ExecuteCommand() *cobra.Command {
	executeCmd.Flags().StringArrayVarP(&resourcePaths, "resource", "r", []string{}, "Path to resource files")
	executeCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Specifies whether to write reports to filesystem")

	return executeCmd
}

var outDirectory string

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "generate",
	Example: generateHelp,
	Run: func(cmd *cobra.Command, componentDefinitionPaths []string) {
		if len(componentDefinitionPaths) == 0 {
			fmt.Println("Path to the local OSCAL file must be present")
			os.Exit(1)
		}

		err := conductGenerate(componentDefinitionPaths, outDirectory)
		if err != nil {
			os.Exit(1)
		}
	},
}

func GenerateCommand() *cobra.Command {
	generateCmd.Flags().StringVarP(&outDirectory, "out", "o", "./out/", "Path to output kyverno policies")

	return generateCmd
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func conductExecute(componentDefinitionPaths []string, resourcePaths []string, dryRun bool) error {
	applyCommandConfig := &ApplyCommandConfig{}
	if len(resourcePaths) > 0 {
		applyCommandConfig.Cluster = false
	} else {
		applyCommandConfig.Cluster = true
	}
	// unmarshall all documents to types.OscalComponentDocument into a slice of component documents
	// Declare empty slice of oscalComponentDocuments
	oscalComponentDefinitions, err := oscalComponentDefinitionsFromPaths(componentDefinitionPaths)
	check(err)

	var complianceReports []types.ComplianceReport

	// with an array/slice of oscalComponentDocuments, search each implemented requirement for a props.name/value (harcoded and specific value for now)
	// copy struct to slice of implementedRequirements
	// foreach oscalCompoentDocument -- foreach implemented requirement -- if props.name == compliance validator
	implementedReqs, err := getImplementedReqs(oscalComponentDefinitions)
	for _, implementedReq := range implementedReqs {

		path, err := generatePolicy(implementedReq, "./")
		if err != nil {
			log.Log.Error(err, "error string")
		}

		if path == "" {
			continue
		}

		// For right now, we are just passing in a single path/policy as we can use the applyCommandHelper function to provide results for parsing
		// We don't want to pass multiple controls at once currently - as the command will aggregate findings and be unable to decipher between
		// when a control passes or fails individually?
		applyCommandConfig.PolicyPaths = []string{path}
		rc, _, _, _, err := applyCommandConfig.applyCommandHelper()
		if err != nil {
			return err
		}
		// Cleanup the policies - TODO: Introduce a flag that retains policies under a new directory
		os.Remove(path)
		var currentReport types.ComplianceReport

		var result string
		if rc.Pass > 0 && rc.Fail == 0 {
			result = "Pass"
		} else {
			result = "Fail"
		}

		currentReport.SourceRequirements = implementedReq
		currentReport.Result = result

		complianceReports = append(complianceReports, currentReport)

		fmt.Printf("UUID: %v\n\tResources Passing: %v\n\tResources Failing: %v\n\tStatus: %v\n", implementedReq.UUID, rc.Pass, rc.Fail, result)
	}
	if err != nil {
		log.Log.Error(err, "error string")
	}

	if !dryRun {
		generateReport(complianceReports)
	}

	return nil
}

func conductGenerate(componentDefinitionPaths []string, outDirectory string) error {
	err := os.MkdirAll(outDirectory, 0755)
	if err != nil {
		logrus.Error("error creating output directory: ", err)
		return err
	}

	oscalComponentDefinitions, err := oscalComponentDefinitionsFromPaths(componentDefinitionPaths)
	if err != nil {
		logrus.Error("error getting oscal component definitions: ", err)
		return err
	}

	implementedReqs, err := getImplementedReqs(oscalComponentDefinitions)
	if err != nil {
		logrus.Error("error getting implemented requirements: ", err)
		return err
	}

	for _, implementedReq := range implementedReqs {
		_, err := generatePolicy(implementedReq, outDirectory)
		if err != nil {
			logrus.Error("error generating policy: ", err)
			return err
		}
	}
	return nil
}

// Open files and attempt to unmarshall to oscal component definition structs
func oscalComponentDefinitionsFromPaths(filepaths []string) (oscalComponentDefinitions []types.OscalComponentDefinition, err error) {
	for _, path := range filepaths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Printf("Path: %v does not exist - unable to digest document\n", path)
			continue
		}

		rawDoc, err := os.ReadFile(path)
		check(err)

		var oscalComponentDefinition types.OscalComponentDefinition

		jsonDoc, err := yaml2.YAMLToJSON(rawDoc)
		if err != nil {
			fmt.Printf("Error converting YAML to JSON: %s\n", err.Error())
		}

		err = json.Unmarshal(jsonDoc, &oscalComponentDefinition)
		check(err)

		oscalComponentDefinitions = append(oscalComponentDefinitions, oscalComponentDefinition)
	}

	return
}

// Parse the ingested documents (POC = 1) for applicable information
// Knowns = this will be a yaml file
// return a slice of Control objects
func getImplementedReqs(componentDefinitions []types.OscalComponentDefinition) (implementedReqs []types.ImplementedRequirementsCustom, err error) {
	for _, componentDefinition := range componentDefinitions {
		for _, component := range componentDefinition.ComponentDefinition.Components {
			for _, controlImplementation := range component.ControlImplementations {
				implementedReqs = append(implementedReqs, controlImplementation.ImplementedRequirements...)
			}
		}
	}
	return
}

// Turn a ruleset into a ClusterPolicy resource for ability to use Kyverno applyCommandHelper without modification
// This needs to copy the rules into a Cluster Policy resource (yaml) and write to individual files
// Kyverno will perform applying these and generating pass/fail results
func generatePolicy(implementedRequirement types.ImplementedRequirementsCustom, outDir string) (policyPath string, err error) {

	if len(implementedRequirement.Rules) != 0 {
		// fmt.Printf("%v", implementedRequirement.Rules[0].Validation.RawPattern)
		policyName := strings.ToLower(implementedRequirement.UUID)
		policy := v1.ClusterPolicy{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterPolicy",
				APIVersion: "kyverno.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: policyName,
			},
			Spec: v1.Spec{
				Rules: implementedRequirement.Rules,
			}}
		// Requires the use of sigs.k8s.io/yaml for proper marshalling
		yamlData, err := yaml1.Marshal(&policy)

		if err != nil {
			fmt.Printf("Error while Marshaling. %v", err)
		}

		fileName := policyName + ".yaml"
		policyPath = path.Join(outDir, fileName)
		err = ioutil.WriteFile(policyPath, yamlData, 0644)
		if err != nil {
			logrus.Error(err)
			panic("Unable to write data into the file")
		}
	}

	return
}

// This is the OSCAL document generation for final output.
// This should include some ability to consolidate controls met in multiple input documents under single control entries
// This should include fields that reference the source of the control to the original document ingested
func generateReport(compiledReport []types.ComplianceReport) (err error) {
	reportData, err := yaml1.Marshal(&compiledReport)
	check(err)

	currentTime := time.Now()
	fileName := "compliance_report-" + currentTime.Format("01-02-2006-15:04:05") + ".yaml"

	err = ioutil.WriteFile(fileName, reportData, 0644)
	check(err)
	return
}

// github.com/kyverno/kyverno v1.9.0 (Copy/Paste - No modification)
func (c *ApplyCommandConfig) applyCommandHelper() (rc *common.ResultCounts, resources []*unstructured.Unstructured, skipInvalidPolicies SkippedInvalidPolicies, pvInfos []common.Info, err error) {
	store.SetMock(true)
	store.SetRegistryAccess(c.RegistryAccess)
	if c.Cluster {
		store.AllowApiCall(true)
	}
	fs := memfs.New()

	if c.ValuesFile != "" && c.VariablesString != "" {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("pass the values either using set flag or values_file flag", err)
	}

	variables, globalValMap, valuesMap, namespaceSelectorMap, subresources, err := common.GetVariable(c.VariablesString, c.ValuesFile, fs, false, "")
	if err != nil {
		if !sanitizederror.IsErrorSanitized(err) {
			return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("failed to decode yaml", err)
		}
		return rc, resources, skipInvalidPolicies, pvInfos, err
	}

	openApiManager, err := openapi.NewManager()
	if err != nil {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("failed to initialize openAPIController", err)
	}

	var dClient dclient.Interface
	if c.Cluster {
		restConfig, err := config.CreateClientConfigWithContext(c.KubeConfig, c.Context)
		if err != nil {
			return rc, resources, skipInvalidPolicies, pvInfos, err
		}
		kubeClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return rc, resources, skipInvalidPolicies, pvInfos, err
		}
		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return rc, resources, skipInvalidPolicies, pvInfos, err
		}
		dClient, err = dclient.NewClient(context.Background(), dynamicClient, kubeClient, 15*time.Minute)
		if err != nil {
			return rc, resources, skipInvalidPolicies, pvInfos, err
		}
	}

	if len(c.PolicyPaths) == 0 {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("require policy", err)
	}

	if (len(c.PolicyPaths) > 0 && c.PolicyPaths[0] == "-") && len(c.ResourcePaths) > 0 && c.ResourcePaths[0] == "-" {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("a stdin pipe can be used for either policies or resources, not both", err)
	}

	var policies []kyvernov1.PolicyInterface

	isGit := common.IsGitSourcePath(c.PolicyPaths)

	if isGit {
		gitSourceURL, err := url.Parse(c.PolicyPaths[0])
		if err != nil {
			fmt.Printf("Error: failed to load policies\nCause: %s\n", err)
			osExit(1)
		}

		pathElems := strings.Split(gitSourceURL.Path[1:], "/")
		if len(pathElems) <= 1 {
			err := fmt.Errorf("invalid URL path %s - expected https://<any_git_source_domain>/:owner/:repository/:branch (without --git-branch flag) OR https://<any_git_source_domain>/:owner/:repository/:directory (with --git-branch flag)", gitSourceURL.Path)
			fmt.Printf("Error: failed to parse URL \nCause: %s\n", err)
			osExit(1)
		}

		gitSourceURL.Path = strings.Join([]string{pathElems[0], pathElems[1]}, "/")
		repoURL := gitSourceURL.String()
		var gitPathToYamls string
		c.GitBranch, gitPathToYamls = common.GetGitBranchOrPolicyPaths(c.GitBranch, repoURL, c.PolicyPaths)
		_, cloneErr := gitutils.Clone(repoURL, fs, c.GitBranch)
		if cloneErr != nil {
			fmt.Printf("Error: failed to clone repository \nCause: %s\n", cloneErr)
			log.Log.V(3).Info(fmt.Sprintf("failed to clone repository  %v as it is not valid", repoURL), "error", cloneErr)
			osExit(1)
		}
		policyYamls, err := gitutils.ListYamls(fs, gitPathToYamls)
		if err != nil {
			return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("failed to list YAMLs in repository", err)
		}
		sort.Strings(policyYamls)
		c.PolicyPaths = policyYamls
	}
	policies, err = common.GetPoliciesFromPaths(fs, c.PolicyPaths, isGit, "")
	if err != nil {
		fmt.Printf("Error: failed to load policies\nCause: %s\n", err)
		osExit(1)
	}

	if len(c.ResourcePaths) == 0 && !c.Cluster {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("resource file(s) or cluster required", err)
	}

	mutateLogPathIsDir, err := checkMutateLogPath(c.MutateLogPath)
	if err != nil {
		if !sanitizederror.IsErrorSanitized(err) {
			return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("failed to create file/folder", err)
		}
		return rc, resources, skipInvalidPolicies, pvInfos, err
	}

	// empty the previous contents of the file just in case if the file already existed before with some content(so as to perform overwrites)
	// the truncation of files for the case when mutateLogPath is dir, is handled under pkg/kyverno/apply/common.go
	if !mutateLogPathIsDir && c.MutateLogPath != "" {
		c.MutateLogPath = filepath.Clean(c.MutateLogPath)
		// Necessary for us to include the file via variable as it is part of the CLI.
		_, err := os.OpenFile(c.MutateLogPath, os.O_TRUNC|os.O_WRONLY, 0o600) // #nosec G304
		if err != nil {
			if !sanitizederror.IsErrorSanitized(err) {
				return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("failed to truncate the existing file at "+c.MutateLogPath, err)
			}
			return rc, resources, skipInvalidPolicies, pvInfos, err
		}
	}

	err = common.PrintMutatedPolicy(policies)
	if err != nil {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("failed to marshal mutated policy", err)
	}

	resources, err = common.GetResourceAccordingToResourcePath(fs, c.ResourcePaths, c.Cluster, policies, dClient, c.Namespace, c.PolicyReport, false, "")
	if err != nil {
		fmt.Printf("Error: failed to load resources\nCause: %s\n", err)
		osExit(1)
	}

	if (len(resources) > 1 || len(policies) > 1) && c.VariablesString != "" {
		return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError("currently `set` flag supports variable for single policy applied on single resource ", nil)
	}

	// get the user info as request info from a different file
	var userInfo v1beta1.RequestInfo
	var subjectInfo store.Subject
	if c.UserInfoPath != "" {
		userInfo, subjectInfo, err = common.GetUserInfoFromPath(fs, c.UserInfoPath, false, "")
		if err != nil {
			fmt.Printf("Error: failed to load request info\nCause: %s\n", err)
			osExit(1)
		}
		store.SetSubjects(subjectInfo)
	}

	if c.VariablesString != "" {
		variables = common.SetInStoreContext(policies, variables)
	}

	var policyRulesCount, mutatedPolicyRulesCount int
	for _, policy := range policies {
		policyRulesCount += len(policy.GetSpec().Rules)
	}

	for _, policy := range policies {
		mutatedPolicyRulesCount += len(policy.GetSpec().Rules)
	}

	msgPolicyRules := "1 policy rule"
	if policyRulesCount > 1 {
		msgPolicyRules = fmt.Sprintf("%d policy rules", policyRulesCount)
	}

	if mutatedPolicyRulesCount > policyRulesCount {
		msgPolicyRules = fmt.Sprintf("%d policy rules", mutatedPolicyRulesCount)
	}

	msgResources := "1 resource"
	if len(resources) > 1 {
		msgResources = fmt.Sprintf("%d resources", len(resources))
	}

	if len(policies) > 0 && len(resources) > 0 {
		if !c.Stdin {
			if mutatedPolicyRulesCount > policyRulesCount {
				fmt.Printf("\nauto-generated pod policies\nApplying %s to %s...\n", msgPolicyRules, msgResources)
			} else {
				fmt.Printf("\nApplying %s to %s...\n", msgPolicyRules, msgResources)
			}
		}
	}

	rc = &common.ResultCounts{}
	skipInvalidPolicies.skipped = make([]string, 0)
	skipInvalidPolicies.invalid = make([]string, 0)

	for _, policy := range policies {
		_, err := policy2.Validate(policy, nil, true, openApiManager)
		if err != nil {
			log.Log.Error(err, "policy validation error")
			if strings.HasPrefix(err.Error(), "variable 'element.name'") {
				skipInvalidPolicies.invalid = append(skipInvalidPolicies.invalid, policy.GetName())
			} else {
				skipInvalidPolicies.skipped = append(skipInvalidPolicies.skipped, policy.GetName())
			}

			continue
		}

		matches := common.HasVariables(policy)
		variable := common.RemoveDuplicateAndObjectVariables(matches)
		if len(variable) > 0 {
			if len(variables) == 0 {
				// check policy in variable file
				if c.ValuesFile == "" || valuesMap[policy.GetName()] == nil {
					skipInvalidPolicies.skipped = append(skipInvalidPolicies.skipped, policy.GetName())
					continue
				}
			}
		}

		kindOnwhichPolicyIsApplied := common.GetKindsFromPolicy(policy, subresources, dClient)

		for _, resource := range resources {
			thisPolicyResourceValues, err := common.CheckVariableForPolicy(valuesMap, globalValMap, policy.GetName(), resource.GetName(), resource.GetKind(), variables, kindOnwhichPolicyIsApplied, variable)
			if err != nil {
				return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError(fmt.Sprintf("policy `%s` have variables. pass the values for the variables for resource `%s` using set/values_file flag", policy.GetName(), resource.GetName()), err)
			}
			applyPolicyConfig := common.ApplyPolicyConfig{
				Policy:               policy,
				Resource:             resource,
				MutateLogPath:        c.MutateLogPath,
				MutateLogPathIsDir:   mutateLogPathIsDir,
				Variables:            thisPolicyResourceValues,
				UserInfo:             userInfo,
				PolicyReport:         c.PolicyReport,
				NamespaceSelectorMap: namespaceSelectorMap,
				Stdin:                c.Stdin,
				Rc:                   rc,
				PrintPatchResource:   true,
				Client:               dClient,
				AuditWarn:            c.AuditWarn,
				Subresources:         subresources,
			}
			_, info, err := common.ApplyPolicyOnResource(applyPolicyConfig)
			if err != nil {
				return rc, resources, skipInvalidPolicies, pvInfos, sanitizederror.NewWithError(fmt.Errorf("failed to apply policy %v on resource %v", policy.GetName(), resource.GetName()).Error(), err)
			}
			pvInfos = append(pvInfos, info)
		}
	}

	return rc, resources, skipInvalidPolicies, pvInfos, nil
}

// checkMutateLogPath - checking path for printing mutated resource (-o flag)
func checkMutateLogPath(mutateLogPath string) (mutateLogPathIsDir bool, err error) {
	if mutateLogPath != "" {
		spath := strings.Split(mutateLogPath, "/")
		sfileName := strings.Split(spath[len(spath)-1], ".")
		if sfileName[len(sfileName)-1] == "yml" || sfileName[len(sfileName)-1] == "yaml" {
			mutateLogPathIsDir = false
		} else {
			mutateLogPathIsDir = true
		}

		err := createFileOrFolder(mutateLogPath, mutateLogPathIsDir)
		if err != nil {
			if !sanitizederror.IsErrorSanitized(err) {
				return mutateLogPathIsDir, sanitizederror.NewWithError("failed to create file/folder.", err)
			}
			return mutateLogPathIsDir, err
		}
	}
	return mutateLogPathIsDir, err
}

// createFileOrFolder - creating file or folder according to path provided
func createFileOrFolder(mutateLogPath string, mutateLogPathIsDir bool) error {
	mutateLogPath = filepath.Clean(mutateLogPath)
	_, err := os.Stat(mutateLogPath)

	if err != nil {
		if os.IsNotExist(err) {
			if !mutateLogPathIsDir {
				// check the folder existence, then create the file
				var folderPath string
				s := strings.Split(mutateLogPath, "/")

				if len(s) > 1 {
					folderPath = mutateLogPath[:len(mutateLogPath)-len(s[len(s)-1])-1]
					_, err := os.Stat(folderPath)
					if os.IsNotExist(err) {
						errDir := os.MkdirAll(folderPath, 0750)
						if errDir != nil {
							return sanitizederror.NewWithError("failed to create directory", err)
						}
					}
				}

				mutateLogPath = filepath.Clean(mutateLogPath)
				// Necessary for us to create the file via variable as it is part of the CLI.
				file, err := os.OpenFile(mutateLogPath, os.O_RDONLY|os.O_CREATE, 0600) // #nosec G304

				if err != nil {
					return sanitizederror.NewWithError("failed to create file", err)
				}

				err = file.Close()
				if err != nil {
					return sanitizederror.NewWithError("failed to close file", err)
				}

			} else {
				errDir := os.MkdirAll(mutateLogPath, 0750)
				if errDir != nil {
					return sanitizederror.NewWithError("failed to create directory", err)
				}
			}

		} else {
			return sanitizederror.NewWithError("failed to describe file", err)
		}
	}

	return nil
}
