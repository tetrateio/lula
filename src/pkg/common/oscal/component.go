package oscal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"

	"github.com/defenseunicorns/lula/src/pkg/common"
	"github.com/defenseunicorns/lula/src/pkg/message"
)

type Requirement struct {
	ImplementedRequirement *oscalTypes.ImplementedRequirementControlImplementation
	ControlImplementation  *oscalTypes.ControlImplementationSet
}

type selection struct {
	HowMany string
	Choice  []string
}

type parameter struct {
	ID     string
	Label  string
	Select *selection
}

type ComponentDefinition struct {
	Model *oscalTypes.ComponentDefinition
}

func NewComponentDefinition() *ComponentDefinition {
	var compDef ComponentDefinition
	compDef.Model = nil
	return &compDef
}

// Create a new ComponentDefinition from a byte array
func (c *ComponentDefinition) NewModel(data []byte) error {
	model, err := NewOscalModel(data)
	if err != nil {
		return err
	}

	c.Model = model.ComponentDefinition

	return nil
}

// Return the type of the component definition
func (*ComponentDefinition) GetType() string {
	return OSCAL_COMPONENT
}

// Returns the complete OSCAL model with component definition
func (c *ComponentDefinition) GetCompleteModel() *oscalTypes.OscalModels {
	return &oscalTypes.OscalModels{
		ComponentDefinition: c.Model,
	}
}

// MakeDeterministic ensures the relevant elements of the Component Definition are sorted deterministically
func (c *ComponentDefinition) MakeDeterministic() error {
	if c.Model == nil {
		return fmt.Errorf("cannot make nil model deterministic")
	}

	MakeComponentDeterminstic(c.Model)
	return nil
}

// HandleExisting updates the existing Component Defintion if a file is provided
func (c *ComponentDefinition) HandleExisting(path string) error {
	exists, err := common.CheckFileExists(path)
	if err != nil {
		return err
	}
	if exists {
		path = filepath.Clean(path)
		existingFileBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
		compDef := NewComponentDefinition()
		err = compDef.NewModel(existingFileBytes)
		if err != nil {
			return err
		}
		err = MergeComponentDefinitions(compDef.Model, c.Model)
		if err != nil {
			return err
		}
		c.Model = compDef.Model
	}
	return nil
}

// RewritePaths finds all the paths in the component definition relative to the baseDir and updates them to be relative to the newDir
// baseDir and newDir must be absolute paths
func (c *ComponentDefinition) RewritePaths(baseDir string, newDir string) (err error) {
	if c.Model == nil {
		return fmt.Errorf("cannot remap paths, model is nil")
	}

	// Rewrite BackMatter paths
	err = RewritePathsBackMatter(c.Model.BackMatter, baseDir, newDir)
	if err != nil {
		return err
	}

	// Rewrite Metadata paths
	err = RewritePathsMetadata(&c.Model.Metadata, baseDir, newDir)
	if err != nil {
		return err
	}

	// Rewrite ImportComponentDefinitions paths
	if c.Model.ImportComponentDefinitions != nil {
		for i, importCompDef := range *c.Model.ImportComponentDefinitions {
			importCompDef.Href, err = common.RemapPath(importCompDef.Href, baseDir, newDir)
			if err != nil {
				return fmt.Errorf("error remapping path %s: %v", importCompDef.Href, err)
			}
			(*c.Model.ImportComponentDefinitions)[i] = importCompDef
		}
	}

	// Rewrite DefinedComponent paths
	if c.Model.Components != nil {
		for _, component := range *c.Model.Components {
			err = RewritePathsLinks(component.Links, baseDir, newDir)
			if err != nil {
				return err
			}

			err = RewritePathsResponsibleRoles(component.ResponsibleRoles, baseDir, newDir)
			if err != nil {
				return err
			}

			err = RewritePathsControlImplementationSet(component.ControlImplementations, baseDir, newDir)
			if err != nil {
				return err
			}
		}
	}

	// Rewrite Capability paths
	if c.Model.Capabilities != nil {
		for _, capability := range *c.Model.Capabilities {
			err = RewritePathsLinks(capability.Links, baseDir, newDir)
			if err != nil {
				return err
			}

			err = RewritePathsControlImplementationSet(capability.ControlImplementations, baseDir, newDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ResolveImportComponentDefinitions is a function that resolves the import-component-definitions by adding the referenced
// component defintitions into the current component definition and re-writing any paths in the component definition to
// be relative to the importing component's directory
// componentDir must be absolute paths
// TODO: Add templating
func (c *ComponentDefinition) ResolveImportComponentDefinitions(componentDir string) error {
	if c.Model == nil {
		return fmt.Errorf("cannot import component definitions, model is nil")
	}

	if c.Model.ImportComponentDefinitions == nil {
		return nil
	}

	// Add data from each to the current component definition
	for _, importCompDef := range *c.Model.ImportComponentDefinitions {
		// Create a new component definition from the imported component definition Href
		importCompDefHrefAbs, err := filepath.Abs(filepath.Join(componentDir, importCompDef.Href))
		if err != nil {
			return err
		}

		importCompDefHrefAbs = filepath.Clean(importCompDefHrefAbs)

		data, err := os.ReadFile(importCompDefHrefAbs)
		if err != nil {
			return err
		}

		importedComponent := NewComponentDefinition()
		err = importedComponent.NewModel(data)
		if err != nil {
			return err
		}

		// Remap paths in the imported component definition to be relative to the working directory
		err = importedComponent.RewritePaths(filepath.Dir(importCompDefHrefAbs), componentDir)
		if err != nil {
			return err
		}

		// Recursively import any component definitions
		err = importedComponent.ResolveImportComponentDefinitions(componentDir)
		if err != nil {
			return err
		}

		// Merge the imported component definition into the current component definition
		err = MergeComponentDefinitions(c.Model, importedComponent.Model)
		if err != nil {
			return err
		}
	}

	// Clear the imported component definitions
	c.Model.ImportComponentDefinitions = nil

	return nil
}

// MergeVariadicComponentDefinition merges multiple variadic component definitions into a single component definition
func MergeVariadicComponentDefinition(compDefs ...*oscalTypes.ComponentDefinition) (mergedCompDef *oscalTypes.ComponentDefinition, err error) {
	for _, compDef := range compDefs {
		if mergedCompDef == nil {
			mergedCompDef = compDef
		} else {
			err = MergeComponentDefinitions(mergedCompDef, compDef)
			if err != nil {
				return nil, err
			}
		}
	}
	return mergedCompDef, nil
}

// This function should perform a merge of two component-definitions where maintaining the original component-definition is the primary concern.
func MergeComponentDefinitions(original *oscalTypes.ComponentDefinition, latest *oscalTypes.ComponentDefinition) error {
	// Nil check on original and latest
	if original == nil {
		return fmt.Errorf("original component-definition is nil")
	}

	if latest == nil {
		return fmt.Errorf("latest component-definition is nil")
	}

	// merge the component-definition.components
	if original.Components != nil && latest.Components != nil {
		original.Components = mergeDefinedComponents(original.Components, latest.Components)
	} else if original.Components == nil && latest.Components != nil {
		original.Components = latest.Components
	}

	// merge the component-definition.back-matter resources
	if original.BackMatter != nil && latest.BackMatter != nil {
		original.BackMatter = &oscalTypes.BackMatter{
			Resources: mergeResources(original.BackMatter.Resources, latest.BackMatter.Resources),
		}
	} else if original.BackMatter == nil && latest.BackMatter != nil {
		original.BackMatter = latest.BackMatter
	}

	// Artifact will be modified - need to update the timestamp and UUID
	original.Metadata.LastModified = time.Now()
	original.UUID = uuid.NewUUID()

	return nil

}

func mergeDefinedComponents(original *[]oscalTypes.DefinedComponent, latest *[]oscalTypes.DefinedComponent) *[]oscalTypes.DefinedComponent {
	originalMap := make(map[string]oscalTypes.DefinedComponent)

	for _, component := range *original {
		originalMap[component.UUID] = component
	}

	latestMap := make(map[string]oscalTypes.DefinedComponent)

	for _, component := range *latest {
		latestMap[component.UUID] = component
	}

	tempItems := make([]oscalTypes.DefinedComponent, 0)
	for key, value := range latestMap {
		if comp, ok := originalMap[key]; ok {
			// if the component exists - merge & append
			comp = *mergeComponents(&comp, &value)
			tempItems = append(tempItems, comp)
			delete(originalMap, key)
		} else {
			// append the component
			tempItems = append(tempItems, value)
		}
	}

	for _, item := range originalMap {
		tempItems = append(tempItems, item)
	}

	return &tempItems
}

func mergeComponents(original *oscalTypes.DefinedComponent, latest *oscalTypes.DefinedComponent) *oscalTypes.DefinedComponent {
	originalMap := make(map[string]oscalTypes.ControlImplementationSet)

	if original.ControlImplementations != nil {
		for _, item := range *original.ControlImplementations {
			originalMap[item.Source] = item
		}
	}

	latestMap := make(map[string]oscalTypes.ControlImplementationSet)

	if latest.ControlImplementations != nil {
		for _, item := range *latest.ControlImplementations {
			latestMap[item.Source] = item
		}
	}

	tempItems := make([]oscalTypes.ControlImplementationSet, 0)
	for key, value := range latestMap {
		if orig, ok := originalMap[key]; ok {
			// if the control implementation exists - merge & append
			orig = *mergeControlImplementations(&orig, &value)
			tempItems = append(tempItems, orig)
			delete(originalMap, key)
		} else {
			// append the component
			tempItems = append(tempItems, value)
		}
	}

	for _, item := range originalMap {
		tempItems = append(tempItems, item)
	}

	// merge Props without duplicating
	// note: this assumes uniqueness of named prop
	if latest.Props != nil {
		if original.Props == nil {
			original.Props = latest.Props
		} else {
			for _, prop := range *latest.Props {
				UpdateProps(prop.Name, prop.Ns, prop.Value, original.Props)
			}
		}
	}

	original.ControlImplementations = &tempItems
	return original
}

func mergeControlImplementations(original *oscalTypes.ControlImplementationSet, latest *oscalTypes.ControlImplementationSet) *oscalTypes.ControlImplementationSet {
	originalMap := make(map[string]oscalTypes.ImplementedRequirementControlImplementation)

	if original.ImplementedRequirements != nil {
		for _, item := range original.ImplementedRequirements {
			originalMap[item.ControlId] = item
		}
	}
	latestMap := make(map[string]oscalTypes.ImplementedRequirementControlImplementation)

	if latest.ImplementedRequirements != nil {
		for _, item := range latest.ImplementedRequirements {
			latestMap[item.ControlId] = item
		}
	}

	tempItems := make([]oscalTypes.ImplementedRequirementControlImplementation, 0)
	for key, latestImp := range latestMap {
		if orig, ok := originalMap[key]; ok {
			// requirement exists in both - update remarks as this is solely owned by the automation
			orig.Remarks = latestImp.Remarks
			// update the links as another critical field
			if orig.Links != nil && latestImp.Links != nil {
				orig.Links = mergeLinks(*orig.Links, *latestImp.Links)
			} else if orig.Links == nil && latestImp.Links != nil {
				orig.Links = latest.Links
			}

			tempItems = append(tempItems, orig)
			delete(originalMap, key)
		} else {
			// append the component
			tempItems = append(tempItems, latestImp)
		}
	}

	for _, item := range originalMap {
		tempItems = append(tempItems, item)
	}
	// merge Props without duplicating
	// note: this assumes uniqueness of named prop
	if latest.Props != nil {
		if original.Props == nil {
			original.Props = latest.Props
		} else {
			for _, prop := range *latest.Props {
				UpdateProps(prop.Name, prop.Ns, prop.Value, original.Props)
			}
		}
	}
	original.ImplementedRequirements = tempItems
	return original
}

// Merges two arrays of links into a single array
// TODO: account for overriding validations
func mergeLinks(orig []oscalTypes.Link, latest []oscalTypes.Link) *[]oscalTypes.Link {
	result := make([]oscalTypes.Link, 0)

	tempLinks := make(map[string]oscalTypes.Link)
	for _, link := range orig {
		// Both of these are string fields, href is required - resource fragment can help establish uniqueness
		key := fmt.Sprintf("%s%s", link.Href, link.ResourceFragment)
		tempLinks[key] = link
		result = append(result, link)
	}

	for _, link := range latest {
		key := fmt.Sprintf("%s%s", link.Href, link.ResourceFragment)
		// Only append if does not exist
		if _, ok := tempLinks[key]; !ok {
			result = append(result, link)
		}
	}

	return &result
}

// Creates a component-definition from a catalog and identified (or all) controls. Allows for specification of what the content of the remarks section should contain.
func ComponentFromCatalog(command string, source string, catalog *oscalTypes.Catalog, componentTitle string, targetControls []string, targetRemarks []string, framework string) (*ComponentDefinition, error) {
	// store all of the implemented requirements
	implementedRequirements := make([]oscalTypes.ImplementedRequirementControlImplementation, 0)
	var componentDefinition = &oscalTypes.ComponentDefinition{}

	if len(targetControls) == 0 {
		return nil, fmt.Errorf("no controls identified for generation")
	}

	controlsToImplement, err := ResolveCatalogControls(catalog, targetControls, nil)
	if err != nil {
		return nil, err
	}

	if len(controlsToImplement) == 0 {
		return nil, fmt.Errorf("no controls were identified in the catalog from the requirements list: %v\n", targetControls)
	}

	for _, control := range controlsToImplement {
		ir, err := ControlToImplementedRequirement(&control, targetRemarks)
		if err != nil {
			return nil, fmt.Errorf("error creating implemented requirement: %v", err)
		}
		implementedRequirements = append(implementedRequirements, ir)
	}

	for _, id := range targetControls {
		if _, ok := controlsToImplement[id]; !ok {
			message.Debugf("Control %s not found", id)
		}
	}

	props := []oscalTypes.Property{
		{
			Name:  "generation",
			Ns:    LULA_NAMESPACE,
			Value: command,
		},
	}

	if framework != "" {
		prop := oscalTypes.Property{
			Name:  "framework",
			Ns:    LULA_NAMESPACE,
			Value: framework,
		}
		props = append(props, prop)
	}

	componentDefinition.Components = &[]oscalTypes.DefinedComponent{
		{
			UUID:        uuid.NewUUID(),
			Type:        "software",
			Title:       componentTitle,
			Description: "Component Description",
			ControlImplementations: &[]oscalTypes.ControlImplementationSet{
				{
					UUID:                    uuid.NewUUIDWithSource(source),
					Source:                  source,
					ImplementedRequirements: implementedRequirements,
					Description:             "Control Implementation Description",
					Props:                   &props,
				},
			},
		},
	}
	rfc3339Time := time.Now()

	componentDefinition.UUID = uuid.NewUUID()

	componentDefinition.Metadata = oscalTypes.Metadata{
		OscalVersion: OSCAL_VERSION,
		LastModified: rfc3339Time,
		Published:    &rfc3339Time,
		Remarks:      "Lula Generated Component Definition",
		Title:        "Component Title",
		Version:      "0.0.1",
	}

	var compDef ComponentDefinition
	compDef.Model = componentDefinition

	return &compDef, nil

}

// Consume a control - Identify statements - iterate through parts in order to create a description
func ControlToImplementedRequirement(control *oscalTypes.Control, targetRemarks []string) (implementedRequirement oscalTypes.ImplementedRequirementControlImplementation, err error) {
	remarks, err := getControlRemarks(control, targetRemarks)
	if err != nil {
		return implementedRequirement, err
	}

	// assemble implemented-requirements object
	implementedRequirement.Remarks = remarks
	implementedRequirement.Description = "<how the specified control may be implemented if the containing component or capability is instantiated in a system security plan>"
	implementedRequirement.ControlId = control.ID
	implementedRequirement.UUID = uuid.NewUUID()

	return implementedRequirement, nil
}

// Returns a map of the uuid - description of the back-matter resources
func BackMatterToMap(backMatter oscalTypes.BackMatter) (resourceMap map[string]string) {
	resourceMap = make(map[string]string)
	if backMatter.Resources == nil {
		return resourceMap
	}

	for _, resource := range *backMatter.Resources {
		// perform a check to see if the key already exists (meaning duplicitive uuid use)
		_, exists := resourceMap[resource.UUID]
		if exists {
			message.Warnf("Duplicative UUID use detected - Overwriting UUID %s", resource.UUID)
		}

		resourceMap[resource.UUID] = resource.Description
	}
	return resourceMap

}

func ControlImplementationstToRequirementsMap(controlImplementations *[]oscalTypes.ControlImplementationSet) (requirementMap map[string]Requirement) {
	requirementMap = make(map[string]Requirement)

	if controlImplementations != nil {
		for _, controlImplementation := range *controlImplementations {
			for _, requirement := range controlImplementation.ImplementedRequirements {
				requirementMap[requirement.UUID] = Requirement{
					ImplementedRequirement: &requirement,
					ControlImplementation:  &controlImplementation,
				}
			}
		}
	}
	return requirementMap
}

func FilterControlImplementations(componentDefinition *oscalTypes.ComponentDefinition) (controlMap map[string][]oscalTypes.ControlImplementationSet) {
	controlMap = make(map[string][]oscalTypes.ControlImplementationSet)

	if componentDefinition.Components != nil {
		// Build a map[source/framework][]control-implementations
		for _, component := range *componentDefinition.Components {
			if component.ControlImplementations != nil {
				for _, controlImplementation := range *component.ControlImplementations {
					// Using UUID here as the key -> could also be string -> what would we rather the user pass in?
					controlMap[controlImplementation.Source] = append(controlMap[controlImplementation.Source], controlImplementation)
					status, value := GetProp("framework", LULA_NAMESPACE, controlImplementation.Props)
					if status {
						controlMap[value] = append(controlMap[value], controlImplementation)
					}
				}
			}
		}
	}

	return controlMap
}

func ComponentsToMap(componentDefinition *oscalTypes.ComponentDefinition) map[string]*oscalTypes.DefinedComponent {
	components := make(map[string]*oscalTypes.DefinedComponent)

	if componentDefinition != nil && componentDefinition.Components != nil {
		for _, component := range *componentDefinition.Components {
			components[component.UUID] = &component
		}
	}
	return components
}

func MakeComponentDeterminstic(component *oscalTypes.ComponentDefinition) {
	// sort components by title

	if component.Components != nil {
		components := *component.Components
		sort.Slice(components, func(i, j int) bool {
			return components[i].Title < components[j].Title
		})

		// sort control-implementations per component by source
		for _, component := range components {
			if component.ControlImplementations != nil {
				controlImplementations := *component.ControlImplementations
				sort.Slice(controlImplementations, func(i, j int) bool {
					return controlImplementations[i].Source < controlImplementations[j].Source
				})
				// sort implemented-requirements per control-implementation by control-id
				for _, controlImplementation := range controlImplementations {
					implementedRequirements := controlImplementation.ImplementedRequirements
					sort.Slice(implementedRequirements, func(i, j int) bool {
						return implementedRequirements[i].ControlId < implementedRequirements[j].ControlId
					})
				}
			}

		}
		component.Components = &components
	}

	// sort capabilities

	if component.Capabilities != nil {
		capabilities := *component.Capabilities
		sort.Slice(capabilities, func(i, j int) bool {
			return capabilities[i].Name < capabilities[j].Name
		})

		for _, capability := range capabilities {

			if capability.ControlImplementations != nil {
				controlImplementations := *capability.ControlImplementations
				sort.Slice(controlImplementations, func(i, j int) bool {
					return controlImplementations[i].Source < controlImplementations[j].Source
				})
				// sort implemented-requirements per control-implementation by control-id
				for _, controlImplementation := range controlImplementations {
					implementedRequirements := controlImplementation.ImplementedRequirements
					sort.Slice(implementedRequirements, func(i, j int) bool {
						return implementedRequirements[i].ControlId < implementedRequirements[j].ControlId
					})
				}
			}

		}
		component.Capabilities = &capabilities
	}

	// sort backmatter
	if component.BackMatter != nil {
		backmatter := *component.BackMatter
		sortBackMatter(&backmatter)
		component.BackMatter = &backmatter
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Function to allow for recursively adding prose to the description string
func addPart(part *[]oscalTypes.Part, paramMap map[string]parameter, level int) string {

	var result, label string

	if part != nil {
		for _, part := range *part {
			// need to get the label first - unsure if there will ever be more than one?
			if part.Props != nil {
				for _, prop := range *part.Props {
					if prop.Name == "label" {
						label = prop.Value
					}
				}
			}

			var tabs string
			// Indents based on labels
			for i := 0; i < level; i++ {
				tabs += "\t"
			}
			// Trims the whitespace
			prose := strings.TrimSpace(part.Prose)
			if prose == "" {
				result += fmt.Sprintf("%s%s\n", tabs, label)
			} else if strings.Contains(prose, "{{ insert: param,") {
				result += fmt.Sprintf("%s%s %s\n", tabs, label, replaceParams(prose, paramMap, false))
			} else {
				result += fmt.Sprintf("%s%s %s\n", tabs, label, prose)
			}
			if part.Parts != nil {
				result += addPart(part.Parts, paramMap, level+1)
			}

		}
	}

	return result
}

func replaceParams(input string, params map[string]parameter, nested bool) string {
	re := regexp.MustCompile(`{{\s*insert:\s*param,\s*([^}\s]+)\s*}}`)
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		paramName := strings.TrimSpace(re.FindStringSubmatch(match)[1])
		if param, ok := params[paramName]; ok {
			if nested {
				// If we know there is no information that requires prepending
				return param.Label
			} else if param.Select == nil {
				// Standard assignment given a label
				return fmt.Sprintf("[Assignment: organization-defined %s]", param.Label)
			} else {
				// Join many choices into a single item of prose
				prose := fmt.Sprintf("[Selection: (%s) organization-defined", param.Select.HowMany)
				for _, choice := range param.Select.Choice {
					// Handle potential nested parameter use
					if strings.Contains(choice, "{{ insert: param,") {
						replaced := replaceParams(choice, params, true)
						prose += fmt.Sprintf(" %s;", replaced)
					} else {
						prose += fmt.Sprintf(" %s;", choice)
					}
				}
				return fmt.Sprintf("%s]", prose)
			}
		}
		return match
	})
	return result
}
