package hardware

import (
	"golang.org/x/sys/cpu"
)

func IsInitialized() bool {
	return cpu.Initialized
}

func IsX86() bool {
	if !cpu.Initialized {
		return false
	}
	return cpu.X86.HasSSE2
}

func IsARM64() bool {
	if !cpu.Initialized {
		return false
	}
	return cpu.ARM64.HasFP
}
