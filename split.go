package bs

import (
	"context"
	"io"

	"github.com/bobg/hashsplit"
	"google.golang.org/protobuf/proto"
)

// SplitWrite writes the contents of `r` to the blob store `s`,
// splitting the input into a tree of blobs according to `splitter`.
// It returns the ref of the root blob,
// which is a serialized TreeNode.
//
// Splitting is done with the "hashsplitting" technique,
// which finds blob boundaries based on the content of the data
// rather than by position.
// If a new version of the same data is written to the store,
// but with a change,
// only the region of the change will need a new blob;
// the others will be unaffected.
//
// If splitter is nil,
// a default splitter is used that produces chunks that are typically 5-10kb in size.
func SplitWrite(ctx context.Context, s Store, r io.Reader, splitter *hashsplit.Splitter) (Ref, error) {
	if splitter == nil {
		splitter = &hashsplit.Splitter{
			Reset:   true, // xxx ?
			MinSize: 1024, // xxx ?
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	chunks := splitter.Split(ctx, r)
	chunks, errfn := hashsplit.Filter(chunks, func(chunk hashsplit.Chunk) (hashsplit.Chunk, error) {
		ref, _, err := s.Put(ctx, chunk.Bytes)
		if err != nil {
			return chunk, err
		}

		chunk2 := chunk
		chunk2.Bytes = ref[:]
		chunk2.Level /= 2 // xxx ?
		return chunk2, nil
	})

	root := hashsplit.Tree(chunks)
	if err := splitter.E; err != nil {
		return Ref{}, err
	}
	if err := errfn(); err != nil {
		return Ref{}, err
	}

	return splitWrite(ctx, s, root)
}

func splitWrite(ctx context.Context, s Store, n *hashsplit.Node) (Ref, error) {
	tn := &TreeNode{Size: n.Size}
	if len(n.Leaves) > 0 {
		tn.Leaves = n.Leaves
	} else {
		for _, child := range n.Nodes {
			childRef, err := splitWrite(ctx, s, child)
			if err != nil {
				return Ref{}, err
			}
			tn.Nodes = append(tn.Nodes, childRef[:])
		}
	}
	ref, _, err := PutProto(ctx, s, tn)
	return ref, err
}

// SplitRead reads blobs from `g`,
// reassembling the content of the blob tree created with SplitWrite
// and writing it to `w`.
// The ref of the root TreeNode is given by `ref`.
func SplitRead(ctx context.Context, g Getter, ref Ref, w io.Writer) error {
	var tn TreeNode
	err := GetProto(ctx, g, ref, &tn)
	if err != nil {
		return err
	}
	return splitRead(ctx, g, &tn, w)
}

func splitRead(ctx context.Context, g Getter, n *TreeNode, w io.Writer) error {
	if len(n.Leaves) > 0 {
		return splitReadHelper(ctx, g, n.Leaves, func(m []byte) error {
			_, err := w.Write(m)
			return err
		})
	}
	return splitReadHelper(ctx, g, n.Nodes, func(m []byte) error {
		var tn TreeNode
		err := proto.Unmarshal(m, &tn)
		if err != nil {
			return err
		}
		return splitRead(ctx, g, &tn, w)
	})
}

func splitReadHelper(ctx context.Context, g Getter, subrefsBytes [][]byte, do func([]byte) error) error {
	subrefs := make([]Ref, len(subrefsBytes))
	for i, b := range subrefsBytes {
		copy(subrefs[i][:], b)
	}
	blobs, err := g.GetMulti(ctx, subrefs)
	if err != nil {
		return err
	}
	for _, subref := range subrefs {
		f := blobs[subref]
		m, err := f(ctx)
		if err != nil {
			return err
		}
		err = do(m)
		if err != nil {
			return err
		}
	}
	return nil
}
