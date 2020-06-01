package file

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/bobg/bs"
)

var _ bs.Store = &Store{}

type Store struct {
	root string
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) blobpath(ref bs.Ref) string {
	h := hex.EncodeToString(ref[:])
	return filepath.Join(s.root, "blobs", h[:2], h[:4], h)
}

func (s *Store) anchorpath(anchor bs.Anchor) string {
	enc := url.PathEscape(string(anchor))
	el1 := enc
	if len(el1) > 2 {
		el1 = el1[:2]
	}
	el2 := enc
	if len(el2) > 4 {
		el2 = el2[:4]
	}
	return filepath.Join(s.root, "anchors", el1, el2, enc)
}

func (s *Store) Get(_ context.Context, ref bs.Ref) (bs.Blob, error) {
	b, err := ioutil.ReadFile(s.blobpath(ref))
	if os.IsNotExist(err) {
		return nil, bs.ErrNotFound
	}
	return b, err
}

func (s *Store) GetMulti(ctx context.Context, refs []bs.Ref) (bs.GetMultiResult, error) {
	return bs.GetMulti(ctx, s, refs)
}

func (s *Store) GetAnchored(ctx context.Context, anchor bs.Anchor, t time.Time) (bs.Ref, bs.Blob, error) {
	dir := s.anchorpath(anchor)
	entries, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return bs.Zero, nil, bs.ErrNotFound
	}
	if err != nil {
		return bs.Zero, nil, err
	}

	// We might use sort.Search here (since ReadDir returns entries sorted by name),
	// which is O(log N),
	// but we want to be robust in the face of filenames that time.Parse fails to parse,
	// so O(N) it is.
	var best string
	for _, entry := range entries {
		name := entry.Name()
		parsed, err := time.Parse(time.RFC3339Nano, name)
		if err != nil {
			continue
		}
		if parsed.After(t) {
			break
		}
		best = name
	}
	if best == "" {
		return bs.Zero, nil, bs.ErrNotFound
	}

	h, err := ioutil.ReadFile(filepath.Join(dir, best))
	if err != nil {
		return bs.Zero, nil, err
	}
	var ref bs.Ref
	_, err = hex.Decode(ref[:], h)
	if err != nil {
		return bs.Zero, nil, err
	}
	b, err := s.Get(ctx, ref)
	return ref, b, err
}

func (s *Store) Put(_ context.Context, b bs.Blob) (bs.Ref, bool, error) {
	var (
		ref  = b.Ref()
		path = s.blobpath(ref)
		dir  = filepath.Dir(path)
	)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return ref, false, err
	}

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		return ref, false, err
	}

	err = ioutil.WriteFile(path, b, 0644)
	return ref, true, err
}

func (s *Store) PutMulti(ctx context.Context, blobs []bs.Blob) (bs.PutMultiResult, error) {
	return bs.PutMulti(ctx, s, blobs)
}

func (s *Store) PutAnchored(ctx context.Context, b bs.Blob, anchor bs.Anchor, t time.Time) (bs.Ref, bool, error) {
	ref, added, err := s.Put(ctx, b)
	if err != nil {
		return ref, added, err
	}

	dir := s.anchorpath(anchor)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return bs.Zero, false, err
	}

	err = ioutil.WriteFile(
		filepath.Join(dir, t.Format(time.RFC3339Nano)),
		[]byte(hex.EncodeToString(ref[:])),
		0644,
	)
	return ref, added, err
}
