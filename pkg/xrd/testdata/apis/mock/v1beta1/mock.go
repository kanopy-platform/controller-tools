// +groupName=testdata.xplane.io
// +versionName=v1beta1
package mock

import (
	xpapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A MockXRD defines a new CompositeResourceDefinition. The new resource is composed of other composite or managed
// infrastructure resources.
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="ESTABLISHED",type="string",JSONPath=".status.conditions[?(@.type=='Established')].status"
// +kubebuilder:printcolumn:name="OFFERED",type="string",JSONPath=".status.conditions[?(@.type=='Offered')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories=crossplane,shortName=xrd;xrds
// +kubebuilder:claim:singular=foo,kind=foos
// +kubebuilder:defaultcompositionref:name=examplecomp,enforced=true
type MockXRD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MockXRDSpec                                `json:"spec,omitempty"`
	Status xpapiext.CompositeResourceDefinitionStatus `json:"status,omitempty"`
}

type MockXRDSpec struct {
	Thing string `json:"thing"`
	//+optional
	OtherThing string `json:"otherThing"`
}
