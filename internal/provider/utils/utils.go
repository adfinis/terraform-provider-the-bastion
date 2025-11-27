// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

func ToPtr[T any](v T) *T {
	return &v
}
