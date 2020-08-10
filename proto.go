package bs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GetProto reads a blob from a blob store and parses it into the given protocol buffer.
func GetProto(ctx context.Context, g Getter, ref Ref, m proto.Message) error {
	// TODO: check type info?
	b, _, err := g.Get(ctx, ref)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, m)
}

func PutProto(ctx context.Context, s Store, m proto.Message) (Ref, bool, error) {
	typeProto := Type(m)

	var typeRef Ref
	if _, ok := m.(*descriptorpb.DescriptorProto); ok {
		typeRef = TypeTypeRef
	} else {
		var err error

		typeRef, _, err = PutProto(ctx, s, typeProto)
		if err != nil {
			return Ref{}, false, errors.Wrap(err, "storing protobuf type")
		}
	}

	b, err := proto.Marshal(m)
	if err != nil {
		return Ref{}, false, errors.Wrap(err, "marshaling protobuf")
	}

	return s.Put(ctx, b, &typeRef)
}

func Type(m proto.Message) proto.Message {
	return protodesc.ToDescriptorProto(m.ProtoReflect().Descriptor())
}

func TypeBlob(m proto.Message) (Blob, error) {
	t := Type(m)
	b, err := proto.Marshal(t)
	return Blob(b), err
}

func TypeRef(m proto.Message) (Ref, error) {
	b, err := TypeBlob(m)
	if err != nil {
		return Ref{}, err
	}
	return b.Ref(), nil
}

func ProtoRef(m proto.Message) (Ref, error) {
	b, err := proto.Marshal(m)
	if err != nil {
		return Ref{}, err
	}
	return Blob(b).Ref(), nil
}

var (
	TypeTypeRef  Ref
	TypeTypeBlob Blob
)

func init() {
	t := Type(&descriptorpb.DescriptorProto{})

	var err error
	TypeTypeBlob, err = TypeBlob(t)
	if err != nil {
		panic(err)
	}

	TypeTypeRef = TypeTypeBlob.Ref()
}

// Experimental! See:
// https://groups.google.com/forum/?utm_medium=email&utm_source=footer#!msg/protobuf/xRWSIyQ3Qyg/YcuGve18BAAJ.

// DynGetProto retrieves the pb.Blob (typed blob) at ref,
// constructs a new *dynamicpb.Message from the described type,
// and loads the nested Blob into it.
// This serializes the same as the original protobuf
// but is not convertible (or type-assertable) to the original protobuf's Go type.
func DynGetProto(ctx context.Context, g Getter, ref Ref) (*dynamicpb.Message, error) {
	b, typ, err := g.Get(ctx, ref)
	if err != nil {
		return nil, errors.Wrapf(err, "getting %s", ref)
	}

	var dp descriptorpb.DescriptorProto
	err = GetProto(ctx, g, typ, &dp)
	if err != nil {
		return nil, errors.Wrapf(err, "getting descriptor proto at %s", typ)
	}

	md, err := descriptorProtoToMessageDescriptor(&dp)
	if err != nil {
		return nil, errors.Wrapf(err, "manifesting descriptor proto at %s", typ)
	}

	dm := dynamicpb.NewMessage(md)
	err = proto.Unmarshal(b, dm)
	return dm, errors.Wrapf(err, "unmarshaling %s into protobuf manifested from descriptor proto at %s", ref, typ)
}

func descriptorProtoToMessageDescriptor(dp *descriptorpb.DescriptorProto) (protoreflect.MessageDescriptor, error) {
	name := "x"
	f, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{Name: &name, MessageType: []*descriptorpb.DescriptorProto{dp}}}})
	if err != nil {
		return nil, errors.Wrap(err, "creating Files object")
	}
	if n := f.NumFiles(); n != 1 {
		return nil, fmt.Errorf("created Files object has %d files (want 1)", n)
	}

	var md protoreflect.MessageDescriptor
	f.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		mds := fd.Messages()
		if n := mds.Len(); n != 1 {
			err = fmt.Errorf("got %d messages in created Files object (want 1)", n)
			return false
		}
		md = mds.Get(0)
		return true
	})

	return md, err
}
