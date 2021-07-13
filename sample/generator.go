package sample

import "library/v1/pb"

func NewKeyboard() *pb.Keyboard {

	keyboard := &pb.Keyboard{
		Layout: randomKeyboardLayout(),
		Backlit: randomKeyboardBacklist(),
	}
	return keyboard
}

