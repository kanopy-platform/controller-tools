package markers

import (
	"fmt"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/xrd/types"
)

var AllDefinitions = []*definitionWithHelp{
	// Reusing storageversion to map to Referenceable to make moving between XRD and CRD more seamless
	must(markers.MakeDefinition("kubebuilder:storageversion", markers.DescribesType, StorageVersion{})).
		WithHelp(StorageVersion{}.Help()),
	must(markers.MakeDefinition("kubebuilder:claim", markers.DescribesType, Claim{})).
		WithHelp(Claim{}.Help()),
}

// +controllertools:marker:generateHelp:category=XRD
// Claim indicats that the XRD should provide a namespaced claim resource
type Claim struct {
	// Singular is required
	Singular string `marker:"singular"`

	// Plural is optional and will be set to Singular + "s" if left blank
	Plural string `marker:"plural,optional"`

	// ShortNames is optional and will be omitted if empty
	ShortNames []string `marker:"shortNames,optional"`

	// Kind is required
	Kind string `marker:"kind"`

	// ListKind is optional and will be Kind + "List" if not set
	ListKind string `marker:"listKind,optional"`

	// Categories is optionl and will be omitted if empty
	Categories []string `marker:"categories,optional"`
}

func (c Claim) ApplyToXRD(spec *types.XRDSpec, version string) error {

	if spec.ClaimNames == nil {
		spec.ClaimNames = &apiext.CustomResourceDefinitionNames{}
	}

	if c.Singular == "" {
		return fmt.Errorf("singular requried: kubebuilder:claim")

	}
	spec.ClaimNames.Singular = c.Singular

	return nil
}

// +controllertools:marker:generateHelp:category=XRD
// StorageVersion nodes the version of a XRD that should be refereanceable as the storageversion
type StorageVersion struct{}

func (s StorageVersion) ApplyToXRD(spec *types.XRDSpec, version string) error {
	if version == "" {
		// single-version, do nothing
		return nil
	}
	// multi-version
	for i := range spec.Versions {
		ver := &spec.Versions[i]
		if ver.Name != version {
			continue
		}
		ver.Referenceable = true
		break
	}
	return nil
}

type definitionWithHelp struct {
	*markers.Definition
	Help *markers.DefinitionHelp
}

func (d *definitionWithHelp) WithHelp(help *markers.DefinitionHelp) *definitionWithHelp {
	d.Help = help
	return d
}

func (d *definitionWithHelp) Register(reg *markers.Registry) error {
	if err := reg.Register(d.Definition); err != nil {
		return err
	}
	if d.Help != nil {
		reg.AddHelp(d.Definition, d.Help)
	}
	return nil
}

func must(def *markers.Definition, err error) *definitionWithHelp {
	return &definitionWithHelp{
		Definition: markers.Must(def, err),
	}
}

// Register registers all definitions for XRD generation to the given registry.
func Register(reg *markers.Registry) error {
	for _, def := range AllDefinitions {
		if err := def.Register(reg); err != nil {
			return err
		}
	}

	return nil
}
