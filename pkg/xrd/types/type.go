package types

import (
	xpapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type XRD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec XRDSpec `json:"spec,omitempty"`
	//	Status xpapiext.CompositeResourceDefinitionStatus `json:"status,omitempty"`
}

type XRDSpec struct {
	// Group specifies the API group of the defined composite resource.
	// Composite resources are served under `/apis/<group>/...`. Must match the
	// name of the XRD (in the form `<names.plural>.<group>`).
	// +immutable
	Group string `json:"group"`

	// Names specifies the resource and kind names of the defined composite
	// resource.
	// +immutable
	Names extv1.CustomResourceDefinitionNames `json:"names"`

	// ClaimNames specifies the names of an optional composite resource claim.
	// When claim names are specified Crossplane will create a namespaced
	// 'composite resource claim' CRD that corresponds to the defined composite
	// resource. This composite resource claim acts as a namespaced proxy for
	// the composite resource; creating, updating, or deleting the claim will
	// create, update, or delete a corresponding composite resource. You may add
	// claim names to an existing CompositeResourceDefinition, but they cannot
	// be changed or removed once they have been set.
	// +immutable
	// +optional
	ClaimNames *extv1.CustomResourceDefinitionNames `json:"claimNames,omitempty"`

	// ConnectionSecretKeys is the list of keys that will be exposed to the end
	// user of the defined kind.
	// If the list is empty, all keys will be published.
	// +optional
	ConnectionSecretKeys []string `json:"connectionSecretKeys,omitempty"`

	// DefaultCompositionRef refers to the Composition resource that will be used
	// in case no composition selector is given.
	// +optional
	DefaultCompositionRef *xpapiext.CompositionReference `json:"defaultCompositionRef,omitempty"`

	// EnforcedCompositionRef refers to the Composition resource that will be used
	// by all composite instances whose schema is defined by this definition.
	// +optional
	// +immutable
	EnforcedCompositionRef *xpapiext.CompositionReference `json:"enforcedCompositionRef,omitempty"`

	// Versions is the list of all API versions of the defined composite
	// resource. Version names are used to compute the order in which served
	// versions are listed in API discovery. If the version string is
	// "kube-like", it will sort above non "kube-like" version strings, which
	// are ordered lexicographically. "Kube-like" versions start with a "v",
	// then are followed by a number (the major version), then optionally the
	// string "alpha" or "beta" and another number (the minor version). These
	// are sorted first by GA > beta > alpha (where GA is a version with no
	// suffix such as beta or alpha), and then by comparing major version, then
	// minor version. An example sorted list of versions: v10, v2, v1, v11beta2,
	// v10beta3, v3beta1, v12alpha1, v11alpha2, foo1, foo10. Note that all
	// versions must have identical schemas; Crossplane does not currently
	// support conversion between different version schemas.
	Versions []XRDVersion `json:"versions"`
}

type XRDVersion struct {
	// Name of this version, e.g. “v1”, “v2beta1”, etc. Composite resources are
	// served under this version at `/apis/<group>/<version>/...` if `served` is
	// true.
	Name string `json:"name"`

	// Referenceable specifies that this version may be referenced by a
	// Composition in order to configure which resources an XR may be composed
	// of. Exactly one version must be marked as referenceable; all Compositions
	// must target only the referenceable version. The referenceable version
	// must be served.
	Referenceable bool `json:"referenceable"`

	// Served specifies that this version should be served via REST APIs.
	Served bool `json:"served"`

	// The deprecated field specifies that this version is deprecated and should
	// not be used.
	// +optional
	Deprecated *bool `json:"deprecated,omitempty"`

	// DeprecationWarning specifies the message that should be shown to the user
	// when using this version.
	// +optional
	DeprecationWarning *string `json:"deprecationWarning,omitempty"`

	// Schema describes the schema used for validation, pruning, and defaulting
	// of this version of the defined composite resource. Fields required by all
	// composite resources will be injected into this schema automatically, and
	// will override equivalently named fields in this schema. Omitting this
	// schema results in a schema that contains only the fields required by all
	// composite resources.
	// +optional
	Schema *XRValidation `json:"schema,omitempty"`

	// AdditionalPrinterColumns specifies additional columns returned in Table
	// output. If no columns are specified, a single column displaying the age
	// of the custom resource is used. See the following link for details:
	// https://kubernetes.io/docs/reference/using-api/api-concepts/#receiving-resources-as-tables
	// +optional
	AdditionalPrinterColumns []extv1.CustomResourceColumnDefinition `json:"additionalPrinterColumns,omitempty"`
}

type XRValidation struct {
	// OpenAPIV3Schema is the OpenAPI v3 schema to use for validation and
	// pruning.
	// +kubebuilder:pruning:PreserveUnknownFields
	OpenAPIV3Schema *extv1.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
}
