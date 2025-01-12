package Cx

CxPolicy[result] {
	document := input.document
	document[i].kind == "PodSecurityPolicy"
	metadata := document[i].metadata
	spec := document[i].spec
	object.get(spec, "requiredDropCapabilities", "undefined") == "undefined"

	result := {
		"documentId": input.document[i].id,
		"searchKey": sprintf("metadata.name={{%s}}.spec", [metadata.name]),
		"issueType": "MissingAttribute",
		"keyExpectedValue": sprintf("metadata.name={{%s}}.spec.requiredDropCapabilities is defined", [metadata.name]),
		"keyActualValue": sprintf("metadata.name={{%s}}.spec.requiredDropCapabilities is undefined", [metadata.name]),
	}
}
