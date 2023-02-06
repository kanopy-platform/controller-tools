package xrd

import (
	"fmt"
	"sort"
	"strings"

	xpapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/gobuffalo/flect"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type SpecMarker interface {
	ApplyToXRD(xrd *XRDSpec, version string) error
}

type Marker interface {
	ApplyToXRD(xrd *XRD, version string) error
}

type Parser struct {
	*crd.Parser
	XRDefinitons map[schema.GroupKind]XRD
	// packages marks packages as loaded, to avoid re-loading them.
	packages map[*loader.Package]struct{}
}

func (p *Parser) init() {
	//	if p.packages == nil {
	//		p.packages = make(map[*loader.Package]struct{})
	//	}
	//	if p.flattener == nil {
	//		p.flattener = &Flattener{
	//			Parser: p,
	//		}
	//	}
	if p.Schemata == nil {
		p.Schemata = make(map[crd.TypeIdent]apiext.JSONSchemaProps)
	}
	if p.Types == nil {
		p.Types = make(map[crd.TypeIdent]*markers.TypeInfo)
	}
	if p.PackageOverrides == nil {
		p.PackageOverrides = make(map[string]crd.PackageOverride)
	}
	if p.GroupVersions == nil {
		p.GroupVersions = make(map[*loader.Package]schema.GroupVersion)
	}
	if p.CustomResourceDefinitions == nil {
		p.CustomResourceDefinitions = make(map[schema.GroupKind]apiext.CustomResourceDefinition)
	}
	if p.FlattenedSchemata == nil {
		p.FlattenedSchemata = make(map[crd.TypeIdent]apiext.JSONSchemaProps)
	}

	if p.XRDefinitons == nil {
		p.XRDefinitons = make(map[schema.GroupKind]XRD)
	}
}

func (p *Parser) NeedXRDFor(groupKind schema.GroupKind, maxDescLen *int) {
	p.init()
	if _, exists := p.CustomResourceDefinitions[groupKind]; exists {
		return
	}
	var packages []*loader.Package
	for pkg, gv := range p.GroupVersions {
		if gv.Group != groupKind.Group {
			continue
		}
		packages = append(packages, pkg)
	}
	defaultPlural := strings.ToLower(flect.Pluralize(groupKind.Kind))
	xrd := XRD{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultPlural + "." + groupKind.Group,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: xpapiext.SchemeGroupVersion.String(),
			Kind:       xpapiext.CompositeResourceDefinitionKind,
		},
		Spec: XRDSpec{
			Names: apiext.CustomResourceDefinitionNames{
				Kind:     groupKind.Kind,
				ListKind: groupKind.Kind + "List",
				Plural:   defaultPlural,
				Singular: strings.ToLower(groupKind.Kind),
			},
		},
	}

	for _, pkg := range packages {
		typeIdent := crd.TypeIdent{Package: pkg, Name: groupKind.Kind}
		typeInfo := p.Types[typeIdent]
		if typeInfo == nil {
			continue
		}
		p.NeedFlattenedSchemaFor(typeIdent)
		fullSchema := p.FlattenedSchemata[typeIdent]
		fullSchema = *fullSchema.DeepCopy() // don't mutate the cache (we might be truncating description, etc)
		if maxDescLen != nil {
			crd.TruncateDescription(&fullSchema, *maxDescLen)
		}

		version := XRDVersion{
			Name: p.GroupVersions[pkg].Version,
			Schema: &XRValidation{
				OpenAPIV3Schema: &fullSchema,
			},
		}
		xrd.Spec.Versions = append(xrd.Spec.Versions, version)
	}

	//TODO: marker application should happen here but we don't have the markers defined yet
	// markers are applied *after* initial generation of objects
	for _, pkg := range packages {
		typeIdent := crd.TypeIdent{Package: pkg, Name: groupKind.Kind}
		typeInfo := p.Types[typeIdent]
		if typeInfo == nil {
			continue
		}
		ver := p.GroupVersions[pkg].Version

		for _, markerVals := range typeInfo.Markers {
			for _, val := range markerVals {
				if specMarker, isSpecMarker := val.(SpecMarker); isSpecMarker {
					if err := specMarker.ApplyToXRD(&xrd.Spec, ver); err != nil {
						pkg.AddError(loader.ErrFromNode(err /* an okay guess */, typeInfo.RawSpec))
					}
				} else if xrdMarker, isXRDMarker := val.(Marker); isXRDMarker {
					if err := xrdMarker.ApplyToXRD(&xrd, ver); err != nil {
						pkg.AddError(loader.ErrFromNode(err /* an okay guess */, typeInfo.RawSpec))
					}
				}
			}
		}
	}

	// fix the name if the plural was changed (this is the form the name *has* to take, so no harm in changing it).
	xrd.Name = xrd.Spec.Names.Plural + "." + groupKind.Group

	if len(xrd.Spec.Versions) == 0 {
		return
	}

	// it is necessary to make sure the order of XRD versions in xrd.Spec.Versions is stable and explicitly set crd.Spec.Version.
	// Otherwise, xrd.Spec.Version may point to different XRD versions across different runs.
	sort.Slice(xrd.Spec.Versions, func(i, j int) bool { return xrd.Spec.Versions[i].Name < xrd.Spec.Versions[j].Name })

	// This is configuration validation to ensure we have at least one
	// storage version and a served versions since both are required
	hasStorage := false
	for _, ver := range xrd.Spec.Versions {
		if ver.Referenceable {
			hasStorage = true
			if !ver.Served {
				packages[0].AddError(fmt.Errorf("XRD for %s version %s is referenceable but not served.", groupKind, ver.Name))
			}
			break
		}
	}
	if !hasStorage {
		// just add the error to the first relevant package for this CRD,
		// since there's no specific error location
		packages[0].AddError(fmt.Errorf("CRD for %s has no storage version", groupKind))
	}

	p.XRDefinitons[groupKind] = xrd
}
