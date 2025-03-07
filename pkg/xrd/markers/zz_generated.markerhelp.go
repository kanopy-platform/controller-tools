//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by helpgen. DO NOT EDIT.

package markers

import (
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func (Claim) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "XRD",
		DetailedHelp: markers.DetailedHelp{
			Summary: "indicats that the XRD should provide a namespaced claim resource TODO: these should really be markers on the claim kinds to link them to an XR kind since we will need to render both. For now keeping the types and names and embeds correct is up to the authors. Claims cannot be colocatged in the same library as the XR or they will get generated.",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"Singular": {
				Summary: "is required",
				Details: "",
			},
			"Plural": {
				Summary: "is optional and will be set to Singular + \"s\" if left blank",
				Details: "",
			},
			"ShortNames": {
				Summary: "is optional and will be omitted if empty",
				Details: "",
			},
			"Kind": {
				Summary: "is required",
				Details: "",
			},
			"ListKind": {
				Summary: "is optional and will be Kind + \"List\" if not set",
				Details: "",
			},
			"Categories": {
				Summary: "is optionl and will be omitted if empty",
				Details: "",
			},
		},
	}
}

func (DefaultCompositionRef) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "XRD",
		DetailedHelp: markers.DetailedHelp{
			Summary: "DefaultCompositonRef - specifies the name of the default comopsition used for the XRD and whether it is enforced name=string enforced=bool",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"Name": {
				Summary: "is required",
				Details: "",
			},
			"Enforced": {
				Summary: "is optional and toggles between DefaultCompositionRef and EnforcedCompositionRef",
				Details: "",
			},
		},
	}
}

func (StorageVersion) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "XRD",
		DetailedHelp: markers.DetailedHelp{
			Summary: "nodes the version of a XRD that should be refereanceable as the storageversion",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{},
	}
}
