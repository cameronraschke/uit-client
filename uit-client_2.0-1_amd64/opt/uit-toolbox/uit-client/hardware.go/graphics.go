package hardware

func HasBuiltInScreen() bool {
	// /sys/class/graphics/fb0
	return false
}

func HasTouchscreen() bool {
	// /sys/class/input/eventX where eventX is a touchscreen device
	return false
}

func HasDedicatedGPU() bool {
	// Check for presence of a dedicated GPU
	return false
}
