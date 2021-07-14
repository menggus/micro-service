package serializer

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
)

// JSON FIle Serializer

func WriteProtobufToJSONFile(message proto.Message, filename string) error {
	data, err := protobufToJSON(message)
	if err != nil {
		return fmt.Errorf("cannot marshal protobuf to json: %w", err)
	}

	err = ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("cannot write file to filename: %w", err)
	}
	return nil
}

func ReadProtobufFromJSONFile(filename string, message proto.Message) error {
	//data, err := ioutil.ReadFile(filename)
	//if err != nil {
	//	return fmt.Errorf("cannot read file from filename: %w", err)
	//}

	//jsonpb.Unmarshal(, message)

	return nil
}

// Binary File Serializer

func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	// proto.message => []byte
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("cannot marshal proto message to binary: %w", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("data cannot write file: %w", err)
	}
	return nil
}

func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read file from filename: %w", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("cannot Unmarshal data to message: %w", err)
	}
	return nil
}
