package xrd

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	xrdmarkers "sigs.k8s.io/controller-tools/pkg/xrd/markers"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/version"
)

// The identifier for v1 CustomResourceDefinitions.
const v1 = "v1"

// The default CustomResourceDefinition version to generate.
const defaultVersion = v1

// +controllertools:marker:generateHelp

// Generator generates CustomResourceDefinition objects.
type Generator struct {
	// IgnoreUnexportedFields indicates that we should skip unexported fields.
	//
	// Left unspecified, the default is false.
	IgnoreUnexportedFields *bool `marker:",optional"`

	// AllowDangerousTypes allows types which are usually omitted from CRD generation
	// because they are not recommended.
	//
	// Currently the following additional types are allowed when this is true:
	// float32
	// float64
	//
	// Left unspecified, the default is false
	AllowDangerousTypes *bool `marker:",optional"`

	// MaxDescLen specifies the maximum description length for fields in CRD's OpenAPI schema.
	//
	// 0 indicates drop the description for all fields completely.
	// n indicates limit the description to at most n characters and truncate the description to
	// closest sentence boundary if it exceeds n characters.
	MaxDescLen *int `marker:",optional"`

	// XRDVersions specifies the target API versions of the CRD type itself to
	// generate. Defaults to v1.
	//
	// Currently, the only supported value is v1.
	//
	// The first version listed will be assumed to be the "default" version and
	// will not get a version suffix in the output filename.
	//
	// You'll need to use "v1" to get support for features like defaulting,
	// along with an API server that supports it (Kubernetes 1.16+).
	XRDVersions []string `marker:"xrdVersions,optional"`

	// GenerateEmbeddedObjectMeta specifies if any embedded ObjectMeta in the CRD should be generated
	GenerateEmbeddedObjectMeta *bool `marker:",optional"`

	// HeaderFile specifies the header text (e.g. license) to prepend to generated files.
	HeaderFile string `marker:",optional"`

	// Year specifies the year to substitute for " YEAR" in the header file.
	Year string `marker:",optional"`
}

func (Generator) CheckFilter() loader.NodeFilter {
	return filterTypesForCRDs
}
func (Generator) RegisterMarkers(into *markers.Registry) error {
	err := crdmarkers.Register(into)
	if err != nil {
		return err
	}
	return xrdmarkers.Register(into)
}

// transformRemoveCRDStatus ensures we do not write the CRD status field.
func transformRemoveCRDStatus(obj map[string]interface{}) error {
	delete(obj, "status")
	return nil
}

func (g Generator) Generate(ctx *genall.GenerationContext) error {
	parser := &Parser{
		Parser: &crd.Parser{
			Collector:                  ctx.Collector,
			Checker:                    ctx.Checker,
			IgnoreUnexportedFields:     true,
			AllowDangerousTypes:        false,
			GenerateEmbeddedObjectMeta: false,
		},
	}

	parser.Collector = ctx.Collector
	parser.Checker = ctx.Checker
	parser.IgnoreUnexportedFields = g.IgnoreUnexportedFields != nil && *g.IgnoreUnexportedFields == true
	parser.AllowDangerousTypes = g.AllowDangerousTypes != nil && *g.AllowDangerousTypes == true
	parser.GenerateEmbeddedObjectMeta = g.GenerateEmbeddedObjectMeta != nil && *g.GenerateEmbeddedObjectMeta == true
	crd.AddKnownTypes(parser.Parser)
	for _, root := range ctx.Roots {
		parser.NeedPackage(root)
	}

	metav1Pkg := crd.FindMetav1(ctx.Roots)
	if metav1Pkg == nil {
		// no objects in the roots, since nothing imported metav1
		return nil
	}

	// TODO: allow selecting a specific object
	kubeKinds := FindKubeKinds(parser.Parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		// no objects in the roots
		return nil
	}

	xrdVersions := g.XRDVersions

	if len(xrdVersions) == 0 {
		xrdVersions = []string{defaultVersion}
	}

	var headerText string

	if g.HeaderFile != "" {
		headerBytes, err := ctx.ReadFile(g.HeaderFile)
		if err != nil {
			return err
		}
		headerText = string(headerBytes)
	}
	headerText = strings.ReplaceAll(headerText, " YEAR", " "+g.Year)

	for _, groupKind := range kubeKinds {
		parser.NeedXRDFor(groupKind, g.MaxDescLen)
		xrdRaw := parser.XRDefinitons[groupKind]

		//addAttribution(&crdRaw)

		// the XRD status should be embedded as a field but we can suppress it in the XRD generation
		// and let crossplane inject it for us
		for _, version := range xrdRaw.Spec.Versions {
			removeXRDStatusProps(version.Schema.OpenAPIV3Schema)
		}

		fileName := fmt.Sprintf("%s_%s.yaml", xrdRaw.Spec.Group, xrdRaw.Spec.Names.Plural)
		if err := ctx.WriteYAML(fileName, headerText, []interface{}{xrdRaw}, genall.WithTransform(transformRemoveXRDStatus)); err != nil {
			return err
		}
	}

	return nil
}

func removeDescriptionFromMetadata(crd *apiext.CustomResourceDefinition) {
	for _, versionSpec := range crd.Spec.Versions {
		if versionSpec.Schema != nil {
			removeDescriptionFromMetadataProps(versionSpec.Schema.OpenAPIV3Schema)
		}
	}
}

func removeDescriptionFromMetadataProps(v *apiext.JSONSchemaProps) {
	if m, ok := v.Properties["metadata"]; ok {
		meta := &m
		if meta.Description != "" {
			meta.Description = ""
			v.Properties["metadata"] = m

		}
	}
}

// FixTopLevelMetadata resets the schema for the top-level metadata field which is needed for CRD validation
func FixTopLevelMetadata(crd apiext.CustomResourceDefinition) {
	for _, v := range crd.Spec.Versions {
		if v.Schema != nil && v.Schema.OpenAPIV3Schema != nil && v.Schema.OpenAPIV3Schema.Properties != nil {
			schemaProperties := v.Schema.OpenAPIV3Schema.Properties
			if _, ok := schemaProperties["metadata"]; ok {
				schemaProperties["metadata"] = apiext.JSONSchemaProps{Type: "object"}
			}
		}
	}
}

// addAttribution adds attribution info to indicate controller-gen tool was used
// to generate this CRD definition along with the version info.
func addAttribution(crd *apiext.CustomResourceDefinition) {
	if crd.ObjectMeta.Annotations == nil {
		crd.ObjectMeta.Annotations = map[string]string{}
	}
	crd.ObjectMeta.Annotations["controller-gen.kubebuilder.io/version"] = version.Version()
}

// FindMetav1 locates the actual package representing metav1 amongst
// the imports of the roots.
func FindMetav1(roots []*loader.Package) *loader.Package {
	for _, root := range roots {
		pkg := root.Imports()["k8s.io/apimachinery/pkg/apis/meta/v1"]
		if pkg != nil {
			return pkg
		}
	}
	return nil
}

// FindKubeKinds locates all types that contain TypeMeta and ObjectMeta
// (and thus may be a Kubernetes object), and returns the corresponding
// group-kinds.
func FindKubeKinds(parser *crd.Parser, metav1Pkg *loader.Package) []schema.GroupKind {
	// TODO(directxman12): technically, we should be finding metav1 per-package
	kubeKinds := map[schema.GroupKind]struct{}{}
	for typeIdent, info := range parser.Types {
		hasObjectMeta := false
		hasTypeMeta := false

		pkg := typeIdent.Package
		pkg.NeedTypesInfo()
		typesInfo := pkg.TypesInfo

		for _, field := range info.Fields {
			if field.Name != "" {
				// type and object meta are embedded,
				// so they can't be this
				continue
			}

			fieldType := typesInfo.TypeOf(field.RawField.Type)
			namedField, isNamed := fieldType.(*types.Named)
			if !isNamed {
				// ObjectMeta and TypeMeta are named types
				continue
			}
			if namedField.Obj().Pkg() == nil {
				// Embedded non-builtin universe type (specifically, it's probably `error`),
				// so it can't be ObjectMeta or TypeMeta
				continue
			}
			fieldPkgPath := loader.NonVendorPath(namedField.Obj().Pkg().Path())
			fieldPkg := pkg.Imports()[fieldPkgPath]

			// Compare the metav1 package by ID and not by the actual instance
			// of the object. The objects in memory could be different due to
			// loading from different root paths, even when they both refer to
			// the same metav1 package.
			if fieldPkg == nil || fieldPkg.ID != metav1Pkg.ID {
				continue
			}

			switch namedField.Obj().Name() {
			case "ObjectMeta":
				hasObjectMeta = true
			case "TypeMeta":
				hasTypeMeta = true
			}
		}

		if !hasObjectMeta || !hasTypeMeta {
			continue
		}

		groupKind := schema.GroupKind{
			Group: parser.GroupVersions[pkg].Group,
			Kind:  typeIdent.Name,
		}
		kubeKinds[groupKind] = struct{}{}
	}

	groupKindList := make([]schema.GroupKind, 0, len(kubeKinds))
	for groupKind := range kubeKinds {
		groupKindList = append(groupKindList, groupKind)
	}
	sort.Slice(groupKindList, func(i, j int) bool {
		return groupKindList[i].String() < groupKindList[j].String()
	})

	return groupKindList
}

// filterTypesForCRDs filters out all nodes that aren't used in CRD generation,
// like interfaces and struct fields without JSON tag.
func filterTypesForCRDs(node ast.Node) bool {
	switch node := node.(type) {
	case *ast.InterfaceType:
		// skip interfaces, we never care about references in them
		return false
	case *ast.StructType:
		return true
	case *ast.Field:
		_, hasTag := loader.ParseAstTag(node.Tag).Lookup("json")
		// fields without JSON tags mean we have custom serialization,
		// so only visit fields with tags.
		return hasTag
	default:
		return true
	}
}

// transformRemoveCRDStatus ensures we do not write the CRD status field.
func transformRemoveXRDStatus(obj map[string]interface{}) error {
	delete(obj, "status")
	return nil
}

func removeXRDStatusProps(v *apiext.JSONSchemaProps) {
	delete(v.Properties, "status")
}
