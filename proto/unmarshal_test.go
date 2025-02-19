package proto

import (
	"fmt"
	"math"
	"testing"

	"github.com/acgreek/protobuf/sample"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUnknownFields(t *testing.T) {
	un := &sample.UnknownFields{
		NumberVal: 8000,
		StringVal: "foo",
		Nested: &sample.UnknownFields_Nested{
			StringVal: "bar",
		},
	}
	output, err := proto.Marshal(un)
	assert.NoError(t, err)
	a := &sample.TestMessage{}
	err = proto.Unmarshal(output, a)
	assert.NoError(t, err)
	b := &sample.TestMessage{}
	err = Unmarshal(output, b)
	assert.NoError(t, err)
	assert.Equal(t, "", cmp.Diff(a, b, cmpopts.IgnoreUnexported(sample.SubMessage{}),
		cmpopts.IgnoreUnexported(sample.TestMessage{})))
	unknownB := b.ProtoReflect().GetUnknown()
	unknownA := b.ProtoReflect().GetUnknown()
	assert.Equal(t, unknownA, unknownB)
}

func TestSimple(t *testing.T) {
	v := &sample.TestMessage{
		NumberVal:  -494,
		StringVal:  "foo",
		FloatNum:   5.43,
		Fixed32Num: 6,
		Fixed64Num: 7,
		DoubleNum:  8.343,
		SubMessage: &sample.SubMessage{
			Name: "foo",
			Id:   22,
		},
		EnumValue:    sample.TestMessage_V1,
		Snumber32Val: -53425,
		Unumber32Val: 53425,
		Snumber64Val: -43425,
		Unumber64Val: 43425,
		Sfixed32Num:  -324,
		Sfixed64Num:  -524,
		BytesVAl:     []byte("hello world"),
		UnpacketInts: []int32{3, 4, 6, 7},
		PacketInts:   []uint32{9, 3, 4, 6, 7},
		SubMessages: []*sample.SubMessage{
			{
				Name: "mary",
				Id:   4,
			},
			{
				Name: "bob",
				Id:   5,
			},
		},
		RepeatedBytes:  [][]byte{[]byte("face"), []byte("head")},
		RepeatedString: []string{"foo", "bar", "baz"},
		PackedSfixed32: []int32{4, 5, -123, 7},
		MaxInt:         math.MaxInt64,
		MinInt:         math.MinInt64,
		MaxUInt:        math.MaxUint64,
		MinUInt:        0,
		MapVal: map[string]sample.TestMessage_FOO{
			"foo": sample.TestMessage_FOO(4),
			"bar": sample.TestMessage_FOO(5),
		},
	}
	output, err := proto.Marshal(v)
	assert.NoError(t, err)
	b := &sample.TestMessage{}
	err = Unmarshal(output, b)
	assert.NoError(t, err)
	assert.Equal(t, v.NumberVal, b.NumberVal)
	assert.Equal(t, v.Fixed32Num, b.Fixed32Num)
	assert.Equal(t, v.Fixed64Num, b.Fixed64Num)
	assert.Equal(t, v.FloatNum, b.FloatNum)
	assert.Equal(t, v.DoubleNum, b.DoubleNum)
	assert.Equal(t, v.EnumValue, b.EnumValue)
	assert.Equal(t, v.Snumber32Val, b.Snumber32Val)
	assert.Equal(t, v.Snumber64Val, b.Snumber64Val)
	assert.Equal(t, v.Unumber32Val, b.Unumber32Val)
	assert.Equal(t, v.Unumber64Val, b.Unumber64Val)
	assert.Equal(t, v.Sfixed32Num, b.Sfixed32Num)
	assert.Equal(t, v.Sfixed64Num, b.Sfixed64Num)
	assert.Equal(t, v.BytesVAl, b.BytesVAl)
	assert.Equal(t, v.UnpacketInts, b.UnpacketInts)
	assert.NotEmpty(t, b.SubMessage)
	if b.SubMessage != nil {
		assert.Equal(t, v.SubMessage.Name, b.SubMessage.Name)
		assert.Equal(t, v.SubMessage.Id, b.SubMessage.Id)
	}
	assert.Equal(t, v.PacketInts, b.PacketInts)
	assert.Equal(t, 2, len(b.SubMessages))
	assert.Equal(t, "mary", b.SubMessages[0].Name)
	assert.Equal(t, "bob", b.SubMessages[1].Name)
	assert.Equal(t, 2, len(b.RepeatedBytes))
	assert.Equal(t, []byte("face"), b.RepeatedBytes[0])
	assert.Equal(t, []byte("head"), b.RepeatedBytes[1])
	assert.Equal(t, 3, len(b.RepeatedString))
	assert.Equal(t, "foo", b.RepeatedString[0])
	assert.Equal(t, "bar", b.RepeatedString[1])
	assert.Equal(t, "baz", b.RepeatedString[2])
	assert.Equal(t, 4, len(b.PackedSfixed32))
	assert.Equal(t, int32(4), b.PackedSfixed32[0])
	assert.Equal(t, int32(5), b.PackedSfixed32[1])
	assert.Equal(t, int32(-123), b.PackedSfixed32[2])
	assert.Equal(t, int32(7), b.PackedSfixed32[3])
	assert.Equal(t, int64(math.MaxInt64), b.MaxInt)
	assert.Equal(t, int64(math.MinInt64), b.MinInt)
	assert.Equal(t, uint64(math.MaxUint64), b.MaxUInt)
	assert.Equal(t, uint64(0), b.MinUInt)
	assert.Equal(t, v.MapVal, b.MapVal)
}
func FuzzUnmarshal(f *testing.F) {
	f.Add([]byte("this really sucks"))
	f.Fuzz(func(t *testing.T, fuzzyData []byte) {
		b := &sample.TestMessage{}
		err := Unmarshal(fuzzyData, b)
		e := &sample.TestMessage{}
		expected := proto.Unmarshal(fuzzyData, e)
		assert.Equal(t, "", cmp.Diff(e, b, cmpopts.IgnoreUnexported(sample.SubMessage{}),
			cmpopts.IgnoreUnexported(sample.TestMessage{})))
		/// if expected == nil && err !=
		assert.Equal(t, err == nil, expected == nil)
	})
}

func TestLargeNumber(t *testing.T) {
	v := &sample.TestMessage{NumberVal: 4534523452, StringVal: "foo"}
	output, err := proto.Marshal(v)
	assert.NoError(t, err)
	fmt.Printf("hello world %v %X\n", v, output)
	err = Unmarshal(output, v)
	assert.NoError(t, err)
}
