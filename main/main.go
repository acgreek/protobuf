package main

import (
	"fmt"
	"os"

	acgreekproto "github.com/acgreek/protobuf/proto"
	"github.com/acgreek/protobuf/sample"
	"google.golang.org/protobuf/proto"
)

func main() {
	v := &sample.TestMessage{
		NumberVal:  4,
		StringVal:  "foo",
		FloatNum:   0.2342,
		Fixed32Num: 32123,
		DoubleNum:  2.34234,
		Fixed64Num: 433242,
		SubMessage: &sample.SubMessage{
			Name: "bar",
			Id:   12,
		},
		MapVal: map[string]sample.TestMessage_FOO{
			"mouse": sample.TestMessage_FOO(32),
			"rat":   sample.TestMessage_FOO(32),
		},
	}
	v.OneOfVal = &sample.TestMessage_OneOfString{"foo"}
	output, err := proto.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hello world %v %X\n", v, output)
		os.Exit(-1)
	}

	fmt.Printf("hello world %v %X\n", v, output)
	err = acgreekproto.Unmarshal(output, v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wire proto failed validation: %v\n", err)
		os.Exit(-1)
	}
	err = proto.Unmarshal(output, v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wire proto failed validation: %v\n", err)
		os.Exit(-1)
	}
}
