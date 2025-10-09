// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ptr creates and returns a pointer to the provided value of any type.
func Ptr[T any](v T) *T { return &v }

// ConvertEnumList converts a list of strings to an enum list of type T.
func ConvertEnumList[T ~int32](raw []string, valueMap map[string]int32) []T {
	result := make([]T, 0, len(raw))
	for _, s := range raw {
		if val, ok := valueMap[s]; ok {
			result = append(result, T(val))
		}
	}
	return result
}

// ExtractStringList extracts a list of strings from a list of types.List.
func ExtractStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) ([]string, bool) {
	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)
	return result, !diags.HasError()
}

// ConvertStringSliceToList converts a []string to types.List.
func ConvertStringSliceToList(strings []string) types.List {
	if len(strings) == 0 {
		return types.ListNull(types.StringType)
	}

	values := make([]attr.Value, 0, len(strings))
	for _, s := range strings {
		values = append(values, types.StringValue(s))
	}
	list, _ := types.ListValue(types.StringType, values)
	return list
}

// ConvertEnumSliceToList converts a slice of protobuf enums to types.List of strings
func ConvertEnumSliceToList[T interface{ String() string }](enums []T) types.List {
	if len(enums) == 0 {
		return types.ListNull(types.StringType)
	}

	values := make([]attr.Value, 0, len(enums))
	for _, e := range enums {
		values = append(values, types.StringValue(e.String()))
	}
	list, _ := types.ListValue(types.StringType, values)
	return list
}
