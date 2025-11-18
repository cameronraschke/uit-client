package hardware

import "golang.org/x/sys/cpu"

func IsARM64() bool {
	return false
}

func IsX86_64() bool {
	return cpu.X86.HasSSE2
}
