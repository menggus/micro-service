package sample

import (
	"library/v1/pb"
	"math/rand"
)

// keyboard
func randomKeyboardLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return pb.Keyboard_QWERTY
	case 2:
		return pb.Keyboard_QWERTZ
	default:
		return pb.Keyboard_AZERTY
	}
}

func randBool() bool {
	// random bool value
	return rand.Intn(2) == 1
}

// cpu
func randomCPUBrand() string {
	return randomStringFromSet("Intel", "Amd")
}

func randomCPUName(brand string) string {
	if brand == "Intel" {
		return randomStringFromSet("Xeon-2286mM", "Core I9-9980HK", "Core I7-9750H", "Core I5-9400F", "Core I3-1005G1")
	}

	return randomStringFromSet("Ryzen 7 PRO 2700U", "Ryzen 5 PRO 3500U", "Ryzen 3 3200GE")
}

// public method random string
func randomStringFromSet(a ...string) string {
	n := len(a)
	if n == 0 {
		return ""
	}
	return a[rand.Intn(n)]
}

// public method random int
func randomInt(min int, max int) int {
	if min > max {
		min, max = max, min
	}
	return min + rand.Intn(max-min)
}

// public method random float64
func randomFloat64(min, max float64) float64 {
	if min > max {
		min, max = max, min
	}
	return min + rand.Float64()*(max-min)
}
