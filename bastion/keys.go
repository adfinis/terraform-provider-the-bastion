// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

// Key represents a Bastion SSH key.
type Key struct {
	Prefix      string   `json:"prefix"`
	ID          string   `json:"id"`
	FromList    []string `json:"fromList"`
	Fingerprint string   `json:"fingerprint"`
	Typecode    string   `json:"typecode"`
	Family      string   `json:"family"`
	Filename    string   `json:"filename"`
	Size        int      `json:"size"`
	Mtime       int      `json:"mtime"`
	Fullpath    string   `json:"fullpath"`
	Comment     string   `json:"comment"`
	Base64      string   `json:"base64"`
	Line        string   `json:"line"`
}

// KeyAlgo represents a Bastion SSH key algorithm.
type KeyAlgo string

const (
	ED25519  KeyAlgo = "ed25519"
	RSA2048  KeyAlgo = "rsa2048"
	RSA4096  KeyAlgo = "rsa4096"
	RSA8192  KeyAlgo = "rsa8192"
	ECDSA256 KeyAlgo = "ecdsa256"
	ECDSA384 KeyAlgo = "ecdsa384"
	ECDSA521 KeyAlgo = "ecdsa521"
)

func (k KeyAlgo) AlgoAndSize() (string, int) {
	switch k {
	case ED25519:
		return "ed25519", 0
	case RSA2048:
		return "rsa", 2048
	case RSA4096:
		return "rsa", 4096
	case RSA8192:
		return "rsa", 8192
	case ECDSA256:
		return "ecdsa", 256
	case ECDSA384:
		return "ecdsa", 384
	case ECDSA521:
		return "ecdsa", 521
	default:
		return "", 0
	}
}
