package markers

import (
	"fmt"

	xpapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
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
	must(markers.MakeDefinition("kubebuilder:defaultcompositionref", markers.DescribesType, DefaultCompositionRef{})).
		WithHelp(DefaultCompositionRef{}.Help()),
}

// +controllertools:marker:generateHelp:category=XRD
// DefaultCompositonRef - specifies the name of the default comopsition
// used for the XRD and whether it is enforced
// name=string
// enforced=bool
type DefaultCompositionRef struct {
	// Name is required
	Name string `marker:"name"`
	// Enforced is optional and toggles between DefaultCompositionRef and EnforcedCompositionRef
	Enforced bool `marker:"enforced,optional"`
}

func (d DefaultCompositionRef) ApplyToXRD(spec *types.XRDSpec, version string) error {
	if spec == nil {
		return nil
	}

	if spec.EnforcedCompositionRef != nil || spec.DefaultCompositionRef != nil {
		return fmt.Errorf("Multiple versions defining CompositionRef settings. Ensure only one struct is marked with defaultcompositionref")
	}

	if d.Name == "" {
		return fmt.Errorf("name requried: kubebuilder:defaultcompositionref:name=<string>,enforced=<bool>")
	}

	compRef := &xpapiext.CompositionReference{
		Name: d.Name,
	}
	if d.Enforced {
		spec.EnforcedCompositionRef = compRef
		spec.DefaultCompositionRef = nil
	} else {
		spec.EnforcedCompositionRef = nil
		spec.DefaultCompositionRef = compRef
	}

	return nil
}

// +controllertools:marker:generateHelp:category=XRD
// Claim indicats that the XRD should provide a namespaced claim resource
// TODO: these should really be markers on the claim kinds to link
// them to an XR kind since we will need to render both. For now
// keeping the types and names and embeds correct is up to the
// authors. Claims cannot be colocatged in the same library as the XR
// or they will get generated.
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
		return fmt.Errorf("singular requried: kubebuilder:claim:singular=<string>,kind=<string>")

	}
	if c.Kind == "" {
		return fmt.Errorf("kind requried: kubebuilder:claim:singular=<string>,kind=<string>")

	}
	spec.ClaimNames.Singular = c.Singular
	spec.ClaimNames.Kind = c.Kind

	spec.ClaimNames.Plural = spec.ClaimNames.Singular + "s"
	if c.Plural != "" {
		spec.ClaimNames.Plural = c.Plural
	}

	spec.ClaimNames.ListKind = spec.ClaimNames.Kind + "List"
	if c.ListKind != "" {
		spec.ClaimNames.ListKind = c.ListKind
	}

	if len(c.ShortNames) > 0 {
		spec.ClaimNames.ShortNames = append(spec.ClaimNames.ShortNames, c.ShortNames...)
	}

	if len(c.Categories) > 0 {
		spec.ClaimNames.Categories = append(spec.ClaimNames.Categories, c.Categories...)
	}

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
