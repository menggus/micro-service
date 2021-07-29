package sample

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"library/v1/pb"
)

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

// NewGPU return a sample GPU
func NewGPU() *pb.GPU {
	memory := &pb.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  pb.Memory_GIGABYTE,
	}
	// cpu
	brand := randomGPUBrand()
	name := randomGPUName(brand)
	minGhz := randomFloat64(2.0, 5.0)
	maxGhz := randomFloat64(2.0, 5.0)
	gpu := &pb.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: memory,
	}

	return gpu
}

// NewRAM return a sample pc ram
func NewRAM() *pb.Memory {
	// memory
	memory := &pb.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  pb.Memory_GIGABYTE,
	}

	return memory
}

// NewStorageSSD return a sample SSD
func NewStorageSSD() *pb.Storage {
	memory := &pb.Memory{
		Value: uint64(randomInt(128, 1024)),
		Unit:  pb.Memory_GIGABYTE,
	}
	ssd := &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: memory,
	}

	return ssd
}

// NewStorageHHD return a sample HHD
func NewStorageHHD() *pb.Storage {
	memory := &pb.Memory{
		Value: uint64(randomInt(1, 6)),
		Unit:  pb.Memory_TERABYTE,
	}
	hhd := &pb.Storage{
		Driver: pb.Storage_HHD,
		Memory: memory,
	}

	return hhd
}

// NewScreen return a sample screen
func NewScreen() *pb.Screen {
	screen := &pb.Screen{
		SizeInch: float32(13),
		Resolution: &pb.Screen_Resolution{
			Height: uint32(14),
			Width:  uint32(16),
		},
		Panel:      pb.Screen_OLED,
		Mutiltouch: randBool(),
	}
	return screen
}

// NewLaptop return a sample laptop
func NewLaptop() *pb.Laptop {
	brand := randomLaptop()
	name := randomLaptopName(brand)
	laptop := &pb.Laptop{
		Id:          uuid.New().String(),
		Brand:       brand,
		Name:        name,
		Cpu:         NewCPU(),
		Ram:         NewRAM(),
		Gpus:        []*pb.GPU{NewGPU()},
		Storages:    []*pb.Storage{NewStorageSSD(), NewStorageHHD()},
		Screen:      NewScreen(),
		Keyboard:    NewKeyboard(),
		Weight:      &pb.Laptop_WeightKg{WeightKg: 1.2},
		PriceUsd:    randomFloat64(2000, 5000),
		ReleaseYear: 2021,
		UpdatedAt:   ptypes.TimestampNow(),
	}

	return laptop
}

// RandomLaptopScore return a random float64 number 1-10
func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
