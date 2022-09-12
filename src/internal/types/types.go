package types

import (
	Kyverno "github.com/kyverno/kyverno/api/kyverno/v1"
	"time"
)

type OscalComponentDefinition struct {
	ComponentDefinition struct {
		UUID     string `yaml:"uuid"`
		Metadata struct {
			Title        string    `yaml:"title"`
			LastModified time.Time `yaml:"last-modified"`
			Version      string    `yaml:"version"`
			OscalVersion string    `yaml:"oscal-version"`
			Parties      []struct {
				UUID  string `yaml:"uuid"`
				Type  string `yaml:"type"`
				Name  string `yaml:"name"`
				Links []struct {
					Href string `yaml:"href"`
					Rel  string `yaml:"rel"`
				} `yaml:"links"`
			} `yaml:"parties"`
		} `yaml:"metadata"`
		Components []struct {
			UUID             string `yaml:"uuid"`
			Type             string `yaml:"type"`
			Title            string `yaml:"title"`
			Description      string `yaml:"description"`
			Purpose          string `yaml:"purpose"`
			ResponsibleRoles []struct {
				RoleID    string `yaml:"role-id"`
				PartyUUID string `yaml:"party-uuid"`
			} `yaml:"responsible-roles"`
			ControlImplementations []struct {
				UUID                    string                          `yaml:"uuid"`
				Source                  string                          `yaml:"source"`
				Description             string                          `yaml:"description"`
				ImplementedRequirements []ImplementedRequirementsCustom `yaml:"implemented-requirements"`
			} `yaml:"control-implementations"`
		}
		BackMatter struct {
			Resources []struct {
				UUID   string `yaml:"uuid"`
				Title  string `yaml:"title"`
				Rlinks []struct {
					Href string `yaml:"href"`
				} `yaml:"rlinks"`
			} `yaml:"resources"`
		} `yaml:"back-matter"`
	} `yaml:"component-definition"`
}

type ImplementedRequirementsCustom struct {
	UUID        string `yaml:"uuid"`
	ControlID   string `yaml:"control-id"`
	Description string `yaml:"description"`
	Rules       []Kyverno.Rule `yaml:"rules,omitempty"`
}
