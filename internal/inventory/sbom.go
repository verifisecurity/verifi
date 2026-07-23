package inventory

import "encoding/json"

// cyclonedx is the subset of the CycloneDX 1.5 spec we emit: enough to describe
// the project and its resolved components with purls.
type cyclonedx struct {
	BOMFormat   string         `json:"bomFormat"`
	SpecVersion string         `json:"specVersion"`
	Version     int            `json:"version"`
	Metadata    cdxMetadata    `json:"metadata"`
	Components  []cdxComponent `json:"components"`
}

type cdxMetadata struct {
	Component cdxComponent `json:"component"`
}

type cdxComponent struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Purl    string `json:"purl,omitempty"`
}

// ToCycloneDX renders the inventory as a CycloneDX 1.5 SBOM (indented JSON).
func (inv *Inventory) ToCycloneDX() ([]byte, error) {
	doc := cyclonedx{
		BOMFormat:   "CycloneDX",
		SpecVersion: "1.5",
		Version:     1,
		Metadata: cdxMetadata{
			Component: cdxComponent{
				Type:    "application",
				Name:    inv.Root.Name,
				Version: inv.Root.Version,
			},
		},
	}
	for _, p := range inv.Packages {
		doc.Components = append(doc.Components, cdxComponent{
			Type:    "library",
			Name:    p.Name,
			Version: p.Version,
			Purl:    p.Purl,
		})
	}
	return json.MarshalIndent(doc, "", "  ")
}
