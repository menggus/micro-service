package serializer

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func ProtobufToJSON(message proto.Message) (string, error) {
	marshaller := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true, // "min_ghz" format in proto file
		//OrigName:     false, // "minGhz" format in *pb.go file
	}

	return marshaller.MarshalToString(message)
}
