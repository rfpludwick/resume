package main

import (
	"log"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	CollapseMultiplePositionsCollapse   = "collapse"
	CollapseMultiplePositionsTitlesOnly = "titles-only"
	CollapseMultiplePositionsFull       = "full"
)

type Configuration struct {
	Controls       ConfigurationControls        `yaml:"controls"`
	Contact        ConfigurationContact         `yaml:"contact"`
	Skills         []ConfigurationSkills        `yaml:"skills"`
	Employment     []ConfigurationOrganization  `yaml:"employment"`
	Volunteering   []ConfigurationOrganization  `yaml:"volunteering"`
	Politics       []ConfigurationOrganization  `yaml:"politics"`
	Education      []ConfigurationEducation     `yaml:"education"`
	Projects       []ConfigurationProject       `yaml:"projects"`
	Certifications []ConfigurationCertification `yaml:"certifications"`
}

type ConfigurationControls struct {
	Pdf            ConfigurationControlsPdf           `yaml:"pdf"`
	Flavor         ConfigurationControlsFlavor        `yaml:"flavor"`
	Skills         ConfigurationControlsSkills        `yaml:"skills"`
	Employers      ConfigurationControlsOrganizations `yaml:"employers"`
	Volunteering   ConfigurationControlsOrganizations `yaml:"volunteering"`
	Politics       ConfigurationControlsOrganizations `yaml:"politics"`
	Education      ConfigurationControlCountTagged    `yaml:"education"`
	Certifications ConfigurationControlCountTagged    `yaml:"certifications"`
}

type ConfigurationControlsPdf struct {
	Filename string                          `yaml:"filename"`
	Fonts    ConfigurationControlsPdfFonts   `yaml:"fonts"`
	Margins  ConfigurationControlsPdfMargins `yaml:"margins"`
	Keywords []string                        `yaml:"keywords"`
}

type ConfigurationControlsPdfFonts struct {
	Header  string `yaml:"header"`
	Footer  string `yaml:"footer"`
	Default string `yaml:"default"`
}

type ConfigurationControlsPdfMargins struct {
	Left  float64 `yaml:"left"`
	Top   float64 `yaml:"top"`
	Right float64 `yaml:"right"`
}

type ConfigurationControlsFlavor struct {
	Header string `yaml:"header"`
	Footer string `yaml:"footer"`
}

type ConfigurationControlsSkills struct {
	First  ConfigurationControlCountTagged `yaml:"first"`
	Second ConfigurationControlCountTagged `yaml:"second"`
	Third  ConfigurationControlCountTagged `yaml:"third"`
}

type ConfigurationControlsOrganizations struct {
	Expanded  ConfigurationControlsOrganizationExpanded `yaml:"expanded"`
	Condensed ConfigurationControlCountTagged           `yaml:"condensed"`
}

type ConfigurationControlsOrganizationExpanded struct {
	Title                     string                                             `yaml:"title"`
	Count                     uint                                               `yaml:"count"`
	BulletPoints              ConfigurationControlsEmployersExpandedBulletPoints `yaml:"bullet_points"`
	CollapseMultiplePositions string                                             `yaml:"collapse_multiple_positions"`
	Tags                      []string                                           `yaml:"tags"`
	PositionTags              []string                                           `yaml:"position_tags"`
}

type ConfigurationControlsEmployersExpandedBulletPoints struct {
	Start     uint `yaml:"start"`
	Decrement uint `yaml:"decrement"`
}

type ConfigurationControlCountTagged struct {
	Title string   `yaml:"title"`
	Count uint     `yaml:"count"`
	Tags  []string `yaml:"tags"`
}

type ConfigurationContact struct {
	Name         string `yaml:"name"`
	EmailAddress string `yaml:"email_address"`
	PhoneNumber  string `yaml:"phone_number"`
	Url          string `yaml:"url"`
	Location     string `yaml:"location"`
}

type ConfigurationSkills struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
	Used bool
}

type ConfigurationOrganization struct {
	Organization      string                              `yaml:"organization"`
	OrganizationExtra string                              `yaml:"organization_extra"`
	Url               string                              `yaml:"url"`
	Location          string                              `yaml:"location"`
	Positions         []ConfigurationOrganizationPosition `yaml:"positions"`
	Tags              []string                            `yaml:"tags"`
	Used              bool
}

type ConfigurationOrganizationPosition struct {
	Title           string             `yaml:"title"`
	NormalizedTitle string             `yaml:"normalized_title"`
	Flavor          string             `yaml:"flavor"`
	Summary         string             `yaml:"summary"`
	Dates           ConfigurationDates `yaml:"dates"`
	BulletPoints    []string           `yaml:"bullet_points,flow"`
	Tags            []string           `yaml:"tags"`
	Used            bool
}

type ConfigurationEducation struct {
	Title       string   `yaml:"title"`
	Url         string   `yaml:"url"`
	Institution string   `yaml:"institution"`
	Tags        []string `yaml:"tags"`
	Used        bool
}

type ConfigurationProject struct {
	Title        string             `yaml:"title"`
	Url          string             `yaml:"url"`
	Location     string             `yaml:"location"`
	Role         string             `yaml:"role"`
	Dates        ConfigurationDates `yaml:"dates"`
	BulletPoints []string           `yaml:"bullet_points,flow"`
	Tags         []string           `yaml:"tags"`
	Used         bool
}

type ConfigurationCertification struct {
	Certification string             `yaml:"certification"`
	Url           string             `yaml:"url"`
	Authority     string             `yaml:"authority"`
	Credential    string             `yaml:"credential"`
	Dates         ConfigurationDates `yaml:"dates"`
	Tags          []string           `yaml:"tags"`
	Used          bool
}

type ConfigurationDates struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

func parseConfiguration() *Configuration {
	// Unmarshal the base resume, secret resume, and controls; in that order they
	// will overwrite the target struct appropriately
	baseResumeFileBody, err := os.ReadFile(flagBaseResumeFile)

	if err != nil {
		log.Fatal("Error reading base resume file:", err)
	}

	var c Configuration

	if err = yaml.Unmarshal(baseResumeFileBody, &c); err != nil {
		log.Fatal("Error decoding base resume YAML:", err)
	}

	secretResumeFileBody, err := os.ReadFile(flagSecretResumeFile)

	if err != nil {
		log.Fatal("Error reading secret resume file:", err)
	}

	if err = yaml.Unmarshal(secretResumeFileBody, &c); err != nil {
		log.Fatal("Error decoding secret resume YAML:", err)
	}

	controlsFileBody, err := os.ReadFile(flagControlsFile)

	if err != nil {
		log.Fatal("Error reading controls file:", err)
	}

	if err = yaml.Unmarshal(controlsFileBody, &c.Controls); err != nil {
		log.Fatal("Error decoding controls YAML:", err)
	}

	// Invalid values (invalues?!) checking
	var validCollapseMultiplePositions = []string{
		CollapseMultiplePositionsCollapse,
		CollapseMultiplePositionsTitlesOnly,
		CollapseMultiplePositionsFull,
	}

	sort.Strings(validCollapseMultiplePositions)

	i := sort.SearchStrings(validCollapseMultiplePositions, c.Controls.Employers.Expanded.CollapseMultiplePositions)

	if (i >= len(validCollapseMultiplePositions)) || (validCollapseMultiplePositions[i] != c.Controls.Employers.Expanded.CollapseMultiplePositions) {
		log.Fatal("Control employers.expanded.collapse_multiple_positions value is invalid: ", c.Controls.Employers.Expanded.CollapseMultiplePositions)
	}

	// Replace newlines with single spaces for expected possible multiline fields
	replacer := strings.NewReplacer(
		"\n\r", " ",
		"\n", " ",
		"\r", " ",
	)

	c.Controls.Flavor.Header = strings.TrimSpace(replacer.Replace(c.Controls.Flavor.Header))
	c.Controls.Flavor.Footer = strings.TrimSpace(replacer.Replace(c.Controls.Flavor.Footer))

	for ei := range c.Employment {
		for pi := range c.Employment[ei].Positions {
			c.Employment[ei].Positions[pi].Summary = strings.TrimSpace(replacer.Replace(c.Employment[ei].Positions[pi].Summary))
			c.Employment[ei].Positions[pi].Flavor = strings.TrimSpace(replacer.Replace(c.Employment[ei].Positions[pi].Flavor))

			for bpi := range c.Employment[ei].Positions[pi].BulletPoints {
				c.Employment[ei].Positions[pi].BulletPoints[bpi] = strings.TrimSpace(replacer.Replace(c.Employment[ei].Positions[pi].BulletPoints[bpi]))
			}
		}
	}

	for vi := range c.Volunteering {
		for pi := range c.Volunteering[vi].Positions {
			c.Volunteering[vi].Positions[pi].Summary = strings.TrimSpace(replacer.Replace(c.Volunteering[vi].Positions[pi].Summary))
			c.Volunteering[vi].Positions[pi].Flavor = strings.TrimSpace(replacer.Replace(c.Volunteering[vi].Positions[pi].Flavor))

			for bpi := range c.Volunteering[vi].Positions[pi].BulletPoints {
				c.Volunteering[vi].Positions[pi].BulletPoints[bpi] = strings.TrimSpace(replacer.Replace(c.Volunteering[vi].Positions[pi].BulletPoints[bpi]))
			}
		}
	}

	for pi := range c.Politics {
		for ppi := range c.Politics[pi].Positions {
			c.Politics[pi].Positions[ppi].Summary = strings.TrimSpace(replacer.Replace(c.Politics[pi].Positions[ppi].Summary))
			c.Politics[pi].Positions[ppi].Flavor = strings.TrimSpace(replacer.Replace(c.Politics[pi].Positions[ppi].Flavor))

			for bpi := range c.Politics[pi].Positions[ppi].BulletPoints {
				c.Politics[pi].Positions[ppi].BulletPoints[bpi] = strings.TrimSpace(replacer.Replace(c.Politics[pi].Positions[ppi].BulletPoints[bpi]))
			}
		}
	}

	return &c
}
