package sample

import "library/v1/pb"

// NewKeyboard return a sample keyboard
func NewKeyboard() *pb.Keyboard {

	keyboard := &pb.Keyboard{
		Layout:  randomKeyboardLayout(), // generate random keyboard layout
		Backlit: randBool(),             // generate random  backlit cato
	}
	return keyboard
}

// NewCPU return a sample CPU
func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)

	numcores := randomInt(2, 8)
	numthreads := randomInt(4, 32)

	minGhz := randomFloat64(2.0, 5.0)
	maxGhz := randomFloat64(2.0, 5.0)

	cpu := &pb.CPU{
		Brand:      brand,
		Name:       name,
		NumCores:   uint32(numcores),
		NumThreads: uint32(numthreads),
		MinGhz:     minGhz,
		MaxGhz:     maxGhz,
	}
	return cpu
}

//
