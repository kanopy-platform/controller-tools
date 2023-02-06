package markers

import (
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/xrd"
)

var XRDMarkers = []*definitionWithHelp{
	// Reusing storageversion to map to Referenceable to make moving between XRD and CRD more seamless
	must(markers.MakeDefinition("kubebuilder:storageversion", markers.DescribesType, StorageVersion{})).
		WithHelp(StorageVersion{}.Help()),
}

type StorageVersions struct{}

func (s StorageVersion) ApplyToXRD(xrd *xrd.XRDSpec, version string) error {
	if version == "" {
		// single-version, do nothing
		return nil
	}
	// multi-version
	for i := range xrd.Spec.Versions {
		ver := &crd.Spec.Versions[i]
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
