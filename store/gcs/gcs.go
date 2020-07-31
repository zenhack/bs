// Package gcs implements a blob store on Google Cloud Storage.
package gcs

import (
	"context"
	"encoding/hex"
	stderrs "errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/bobg/bs"
	"github.com/bobg/bs/store"
)

var _ bs.Store = &Store{}

// Store is a Google Cloud Storage-based implementation of a blob store.
type Store struct {
	bucket *storage.BucketHandle
}

// New produces a new Store.
func New(bucket *storage.BucketHandle) *Store {
	return &Store{bucket: bucket}
}

// Get gets the blob with hash `ref`.
func (s *Store) Get(ctx context.Context, ref bs.Ref) (bs.Blob, error) {
	name := blobObjName(ref)
	obj := s.bucket.Object(name)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "reading info of object %s", name)
	}
	b := make([]byte, r.Size())
	_, err = r.Read(b)
	return b, errors.Wrapf(err, "reading contents of object %s", name)
}

// GetMulti gets multiple blobs in one call.
func (s *Store) GetMulti(ctx context.Context, refs []bs.Ref) (bs.GetMultiResult, error) {
	result := make(bs.GetMultiResult)
	for _, ref := range refs {
		var (
			ref = ref
			ch  = make(chan struct{})
			b   []byte
			err error
		)
		go func() {
			b, err = s.Get(ctx, ref)
			close(ch)
		}()
		result[ref] = func(ctx context.Context) (bs.Blob, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ch:
				return b, err
			}
		}
	}
	return result, nil
}

// GetAnchor gets the latest blob ref for a given anchor as of a given time.
func (s *Store) GetAnchor(ctx context.Context, a bs.Anchor, when time.Time) (bs.Ref, error) {
	var (
		prefix = anchorPrefix(a)
		iter   = s.bucket.Objects(ctx, &storage.Query{Prefix: prefix})
	)

	// Anchors come back in reverse chronological order
	// (since we usually want the latest one).
	// Find the first one whose timestamp is `when` or earlier.
	for {
		attrs, err := iter.Next()
		if stderrs.Is(err, iterator.Done) {
			return bs.Ref{}, bs.ErrNotFound
		}
		if err != nil {
			return bs.Ref{}, errors.Wrap(err, "iterating over anchor objects")
		}
		_, atime, err := anchorTimeFromObjName(attrs.Name)
		if err != nil {
			return bs.Ref{}, errors.Wrapf(err, "decoding object name %s", attrs.Name)
		}
		if atime.After(when) {
			continue
		}
		obj := s.bucket.Object(attrs.Name)
		r, err := obj.NewReader(ctx)
		if err != nil {
			return bs.Ref{}, errors.Wrapf(err, "reading info of object %s", attrs.Name)
		}
		var ref bs.Ref
		if r.Size() != int64(len(ref)) {
			return bs.Ref{}, errors.Wrapf(err, "object %s has wrong size %d (want %d)", attrs.Name, r.Size(), len(ref))
		}
		_, err = r.Read(ref[:])
		return ref, errors.Wrapf(err, "reading contents of object %s", attrs.Name)
	}
}

// Put adds a blob to the store if it wasn't already present.
func (s *Store) Put(ctx context.Context, b bs.Blob) (bs.Ref, bool, error) {
	var (
		ref  = b.Ref()
		name = blobObjName(ref)
		obj  = s.bucket.Object(name).If(storage.Conditions{DoesNotExist: true})
		w    = obj.NewWriter(ctx)
	)
	_, err := w.Write(b) // TODO: are partial writes a possibility?
	var e *googleapi.Error
	if stderrs.As(err, &e) && e.Code == http.StatusPreconditionFailed {
		return ref, false, nil
	}
	if err != nil {
		return ref, false, errors.Wrapf(err, "writing object %s", name)
	}
	err = w.Close()
	return ref, true, errors.Wrapf(err, "closing object %s", name)
}

// PutMulti adds multiple blobs to the store in one call.
func (s *Store) PutMulti(ctx context.Context, blobs []bs.Blob) (bs.PutMultiResult, error) {
	result := make(bs.PutMultiResult, len(blobs))
	for i, b := range blobs {
		var (
			i     = i
			b     = b
			ch    = make(chan struct{})
			ref   bs.Ref
			added bool
			err   error
		)
		go func() {
			ref, added, err = s.Put(ctx, b)
			close(ch)
		}()
		result[i] = func(ctx context.Context) (bs.Ref, bool, error) {
			select {
			case <-ctx.Done():
				return bs.Ref{}, false, ctx.Err()
			case <-ch:
				return ref, added, err
			}
		}
	}
	return result, nil
}

// PutAnchor adds a new ref for a given anchor as of a given time.
func (s *Store) PutAnchor(ctx context.Context, ref bs.Ref, a bs.Anchor, when time.Time) error {
	var (
		name = anchorObjName(a, when)
		obj  = s.bucket.Object(name)
		w    = obj.NewWriter(ctx)
	)
	_, err := w.Write(ref[:])
	if err != nil {
		return errors.Wrap(err, "writing anchor")
	}
	err = w.Close()
	return errors.Wrap(err, "closing anchor")
}

// ListRefs produces all blob refs in the store, in lexical order.
func (s *Store) ListRefs(ctx context.Context, start bs.Ref, f func(bs.Ref) error) error {
	// Google Cloud Storage iterators have no API for starting in the middle of a bucket.
	// But they can filter by object-name prefix.
	// So we take (the hex encoding of) `start` and repeatedly compute prefixes for the objects we want.
	// If `start` is e67a, for example, the sequence of generated prefixes is:
	//   e67b e67c e67d e67e e67f
	//   e68 e69 e6a e6b e6c e6d e6e e6f
	//   e7 e8 e9 ea eb ec ed ee ef
	//   f
	return eachHexPrefix(start.String(), false, func(prefix string) error {
		return s.listRefs(ctx, prefix, f)
	})
}

func (s *Store) listRefs(ctx context.Context, prefix string, f func(bs.Ref) error) error {
	iter := s.bucket.Objects(ctx, &storage.Query{Prefix: "b:" + prefix})
	for {
		obj, err := iter.Next()
		if stderrs.Is(err, iterator.Done) {
			return nil
		}
		if err != nil {
			return err
		}
		ref, err := bs.RefFromHex(obj.Name)
		if err != nil {
			return err
		}
		err = f(ref)
		if err != nil {
			return err
		}
	}
}

// ListAnchors lists all anchors in the store, in lexical order.
func (s *Store) ListAnchors(ctx context.Context, start bs.Anchor, f func(bs.Anchor) error) error {
	startHex := hex.EncodeToString([]byte(start))
	return eachHexPrefix(startHex+"0", true, func(prefix string) error {
		return s.listAnchors(ctx, prefix, f)
	})
}

func (s *Store) listAnchors(ctx context.Context, prefix string, f func(bs.Anchor) error) error {
	iter := s.bucket.Objects(ctx, &storage.Query{Prefix: "a:" + prefix})
	for {
		obj, err := iter.Next()
		if stderrs.Is(err, iterator.Done) {
			return nil
		}
		if err != nil {
			return err
		}
		anchor, err := hex.DecodeString(obj.Name)
		if err != nil {
			return err
		}
		err = f(bs.Anchor(anchor))
		if err != nil {
			return err
		}
	}
}

// ListAnchorRefs lists all blob refs for a given anchor,
// together with their timestamps,
// in chronological order.
func (s *Store) ListAnchorRefs(ctx context.Context, a bs.Anchor, f func(bs.TimeRef) error) error {
	var (
		prefix = anchorPrefix(a)
		iter   = s.bucket.Objects(ctx, &storage.Query{Prefix: prefix})
		pairs  []bs.TimeRef
	)
	for {
		attrs, err := iter.Next()
		if stderrs.Is(err, iterator.Done) {
			break
		}
		got, atime, err := anchorTimeFromObjName(attrs.Name)
		if err != nil {
			return errors.Wrapf(err, "decoding object name %s", attrs.Name)
		}
		if got != a {
			continue
		}
		obj := s.bucket.Object(attrs.Name)
		r, err := obj.NewReader(ctx)
		if err != nil {
			return errors.Wrapf(err, "reading info of object %s", attrs.Name)
		}
		var ref bs.Ref
		if r.Size() != int64(len(ref)) {
			return errors.Wrapf(err, "object %s has wrong size %d (want %d)", attrs.Name, r.Size(), len(ref))
		}
		_, err = r.Read(ref[:])
		if err != nil {
			return errors.Wrapf(err, "reading contents of object %s", attrs.Name)
		}
		pairs = append(pairs, bs.TimeRef{T: atime, R: ref})
	}
	for i := len(pairs) - 1; i >= 0; i-- {
		err := f(pairs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func eachHexPrefix(prefix string, incl bool, f func(string) error) error {
	prefix = strings.ToLower(prefix)
	for len(prefix) > 0 {
		end := hexval(prefix[len(prefix)-1:][0])
		if !incl {
			end++
		}
		prefix = prefix[:len(prefix)-1]
		for c := end; c < 16; c++ {
			err := f(prefix + string(hexdigit(c)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func hexval(b byte) int {
	switch {
	case '0' <= b && b <= '9':
		return int(b - '0')
	case 'a' <= b && b <= 'f':
		return int(10 + b - 'a')
	case 'A' <= b && b <= 'F':
		return int(10 + b - 'A')
	}
	return 0
}

func hexdigit(n int) byte {
	if n < 10 {
		return byte(n + '0')
	}
	return byte(n - 10 + 'a')
}

func blobObjName(ref bs.Ref) string {
	return "b:" + ref.String()
}

func anchorPrefix(a bs.Anchor) string {
	return "a:" + hex.EncodeToString([]byte(a)) + ":"
}

func anchorObjName(a bs.Anchor, when time.Time) string {
	return fmt.Sprintf("%s%020d", anchorPrefix(a), maxTime.Sub(when))
}

var anchorNameRegex = regexp.MustCompile(`^a:([0-9a-f]+):(\d{20})$`)

func anchorTimeFromObjName(name string) (bs.Anchor, time.Time, error) {
	m := anchorNameRegex.FindStringSubmatch(name)
	if len(m) < 3 {
		return "", time.Time{}, errors.New("malformed name")
	}
	a, err := hex.DecodeString(m[1])
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "hex-decoding anchor")
	}
	nanos, err := strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "parsing int64")
	}
	return bs.Anchor(a), maxTime.Add(time.Duration(-nanos)), nil
}

// This is from https://stackoverflow.com/a/32620397
var maxTime = time.Unix(1<<63-1-int64((1969*365+1969/4-1969/100+1969/400)*24*60*60), 999999999)

func init() {
	store.Register("gcs", func(ctx context.Context, conf map[string]interface{}) (bs.Store, error) {
		var options []option.ClientOption
		creds, ok := conf["creds"].(string)
		if !ok {
			return nil, errors.New(`missing "creds" parameter`)
		}
		bucketName, ok := conf["bucket"].(string)
		if !ok {
			return nil, errors.New(`missing "bucket" parameter`)
		}
		options = append(options, option.WithCredentialsFile(creds))
		c, err := storage.NewClient(ctx, options...)
		if err != nil {
			return nil, errors.Wrap(err, "creating cloud storage client")
		}
		return New(c.Bucket(bucketName)), nil
	})
}
