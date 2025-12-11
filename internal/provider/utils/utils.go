// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package utils

func ToPtr[T any](v T) *T {
	return &v
}
