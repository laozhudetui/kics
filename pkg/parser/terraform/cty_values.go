package terraform

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func convertToBoolean(val cty.Value, source byte) (inputVariables, error) {
	var value bool
	if err := gocty.FromCtyValue(val, &value); err != nil {
		return inputVariables{}, err
	}
	return inputVariables{
		value:   value,
		varType: "bool",
		source:  source,
	}, nil
}

func convertToString(val cty.Value, source byte) (inputVariables, error) {
	var value string
	if err := gocty.FromCtyValue(val, &value); err != nil {
		return inputVariables{}, err
	}
	return inputVariables{
		value:   value,
		varType: "string",
		source:  source,
	}, nil
}

func convertToList(val cty.Value, source byte) (inputVariables, error) {
	values := make([]string, 0)
	for _, element := range val.AsValueSlice() {
		var value string
		if err := gocty.FromCtyValue(element, &value); err != nil {
			return inputVariables{}, err
		}
		values = append(values, value)
	}
	return inputVariables{
		value:   values,
		varType: "list",
		source:  source,
	}, nil
}

func convertToMap(val cty.Value, source byte) (inputVariables, error) {
	values := make(map[string]string)
	for key, element := range val.AsValueMap() {
		var value string
		if err := gocty.FromCtyValue(element, &value); err != nil {
			return inputVariables{}, err
		}
		values[key] = value
	}
	return inputVariables{
		value:   values,
		varType: "map",
		source:  source,
	}, nil
}
