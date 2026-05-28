//go:build !(linux && amd64)

package kernel

import "fmt"

func embeddedKernel() ([]byte, error) {
	return nil, fmt.Errorf("embedded kernel only available on linux/amd64; install manually with: mihomo-cli kernel install --local <path>")
}
