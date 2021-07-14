package sample

import (
	"library/v1/pb"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // rand包一般使用固定的种子值，设定每次启动都设置不同的种子值
}

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

// gpu
func randomGPUBrand() string {
	return randomStringFromSet("NVIDIA", "AMD")
}

func randomGPUName(brand string) string {
	if brand == "NVIDIA" {
		return randomStringFromSet(
			"GTX 1070",
			"GTX 1080",
			"RTX 3070",
			"RTX 3080",
		)
	}
	return randomStringFromSet(
		"RX 6800XT",
		"RX 6800",
		"RX 5700",
		"RX 5600",
	)
}

//func randomUnit() pb.Memory_Unit {
//	return pb.Memory_Unit(int32(rand.Intn(6)))
//}
//
//func randomValue(unit pb.Memory_Unit) uint64 {
//	v := rand.Intn(12) // GB
//	switch int(unit) {
//	case 1:
//		return uint64(v * 1024 * 1024 * 1024)
//	case 2:
//		return uint64(v * 1024 * 1024)
//	case 3:
//		return uint64(v * 1024)
//	case 4:
//		return uint64(v)
//	default:
//		return 0
//	}
//}

func randomLaptop() string {

	return randomStringFromSet("Apple", "Dell", "Lenovo")
}

func randomLaptopName(brand string) string {
	switch brand {
	case "Apple":
		return randomStringFromSet("MackBook AIR", "MacBook PRO")
	case "DEll":
		return randomStringFromSet("Latitude", "XPS", "Vostro", "Alienware")
	default:
		return randomStringFromSet("ThinkPad X1", "ThinkPad P1", "ThinkPad P53")
	}
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
