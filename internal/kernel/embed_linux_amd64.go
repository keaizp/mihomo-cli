//go:build linux && amd64

package kernel

import (
	_ "embed"
)

//go:embed embedded/mihomo-linux-amd64.gz
var rawEmbedded []byte

func embeddedKernel() ([]byte, error) {
	return rawEmbedded, nil
}
