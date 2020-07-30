package file

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"hash/adler32"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/bobg/bs"
	"github.com/bobg/bs/store"
)

var _ bs.Store = &Store{}

// Store is a file-based implementation of a blob store.
type Store struct {
	root string
}

// New produces a new Store storing data beneath `root`.
func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) blobroot() string {
	return filepath.Join(s.root, "blobs")
}

func (s *Store) blobpath(ref bs.Ref) string {
	h := ref.String()
	return filepath.Join(s.blobroot(), h[:2], h[:4], h)
}

func (s *Store) anchorroot() string {
	return filepath.Join(s.root, "anchors")
}

func (s *Store) anchorpath(a bs.Anchor) string {
	sum := adler32.Checksum([]byte(a))
	var sumbytes [4]byte
	binary.BigEndian.PutUint32(sumbytes[:], sum)
	sumhex := hex.EncodeToString(sumbytes[:])
	return filepath.Join(s.anchorroot(), sumhex[:2], sumhex, encodeAnchor(a))
}

// Get gets the blob with hash `ref`.
func (s *Store) Get(_ context.Context, ref bs.Ref) (bs.Blob, error) {
	b, err := ioutil.ReadFile(s.blobpath(ref))
	if os.IsNotExist(err) {
		return nil, bs.ErrNotFound
	}
	return b, errors.Wrapf(err, "reading file %s", s.blobpath(ref))
}

// GetMulti gets multiple blobs in one call.
func (s *Store) GetMulti(ctx context.Context, refs []bs.Ref) (bs.GetMultiResult, error) {
	return bs.GetMulti(ctx, s, refs)
}

// GetAnchor gets the latest blob ref for a given anchor as of a given time.
func (s *Store) GetAnchor(ctx context.Context, a bs.Anchor, at time.Time) (bs.Ref, error) {
	dir := s.anchorpath(a)
	entries, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return bs.Ref{}, bs.ErrNotFound
	}
	if err != nil {
		return bs.Ref{}, errors.Wrapf(err, "reading dir %s", dir)
	}

	// We might use sort.Search here (since ReadDir returns entries sorted by name),
	// which is O(log N),
	// but we want to be robust in the face of filenames that time.Parse fails to parse,
	// so O(N) it is.
	// Oh, also: filenames are times that may be expressed in different timezones.
	var best string
	for _, entry := range entries {
		name := entry.Name()
		parsed, err := time.Parse(time.RFC3339Nano, name)
		if err != nil {
			continue
		}
		if parsed.After(at) {
			break
		}
		best = name
	}
	if best == "" {
		return bs.Ref{}, bs.ErrNotFound
	}

	h, err := ioutil.ReadFile(filepath.Join(dir, best))
	if err != nil {
		return bs.Ref{}, errors.Wrapf(err, "reading file %s/%s", dir, best)
	}

	return bs.RefFromHex(string(h))
}

// Put adds a blob to the store if it wasn't already present.
func (s *Store) Put(_ context.Context, b bs.Blob) (bs.Ref, bool, error) {
	var (
		ref  = b.Ref()
		path = s.blobpath(ref)
		dir  = filepath.Dir(path)
	)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return ref, false, errors.Wrapf(err, "ensuring path %s exists", dir)
	}

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		return ref, false, errors.Wrapf(err, "statting %s", path)
	}

	err = ioutil.WriteFile(path, b, 0444)
	return ref, true, errors.Wrapf(err, "writing file %s", path)
}

// PutMulti adds multiple blobs to the store in one call.
func (s *Store) PutMulti(ctx context.Context, blobs []bs.Blob) (bs.PutMultiResult, error) {
	return bs.PutMulti(ctx, s, blobs)
}

// PutAnchor adds a new ref for a given anchor as of a given time.
func (s *Store) PutAnchor(ctx context.Context, ref bs.Ref, a bs.Anchor, at time.Time) error {
	dir := s.anchorpath(a)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return errors.Wrapf(err, "ensuring path %s exists", dir)
	}
	return ioutil.WriteFile(
		filepath.Join(dir, at.Format(time.RFC3339Nano)),
		[]byte(ref.String()),
		0644,
	)
}

// ListRefs produces all blob refs in the store, in lexical order.
func (s *Store) ListRefs(ctx context.Context, start bs.Ref, f func(bs.Ref) error) error {
	topLevel, err := ioutil.ReadDir(s.blobroot())
	if err != nil {
		return errors.Wrapf(err, "reading dir %s", s.blobroot())
	}

	startHex := start.String()
	topIndex := sort.Search(len(topLevel), func(n int) bool {
		return topLevel[n].Name() >= startHex[:2]
	})
	for i := topIndex; i < len(topLevel); i++ {
		topInfo := topLevel[i]
		if !topInfo.IsDir() {
			continue
		}
		topName := topInfo.Name()
		if len(topName) != 2 {
			continue
		}
		if _, err = strconv.ParseInt(topName, 16, 64); err != nil {
			continue
		}

		midLevel, err := ioutil.ReadDir(filepath.Join(s.blobroot(), topName))
		if err != nil {
			return errors.Wrapf(err, "reading dir %s/%s", s.blobroot(), topName)
		}
		midIndex := sort.Search(len(midLevel), func(n int) bool {
			return midLevel[n].Name() >= startHex[:4]
		})
		for j := midIndex; j < len(midLevel); j++ {
			midInfo := midLevel[j]
			if !midInfo.IsDir() {
				continue
			}
			midName := midInfo.Name()
			if len(midName) != 4 {
				continue
			}
			if _, err = strconv.ParseInt(midName, 16, 64); err != nil {
				continue
			}

			blobInfos, err := ioutil.ReadDir(filepath.Join(s.blobroot(), topName, midName))
			if err != nil {
				return errors.Wrapf(err, "reading dir %s/%s/%s", s.blobroot(), topName, midName)
			}

			index := sort.Search(len(blobInfos), func(n int) bool {
				return blobInfos[n].Name() > startHex
			})
			for k := index; k < len(blobInfos); k++ {
				blobInfo := blobInfos[k]
				if blobInfo.IsDir() {
					continue
				}

				ref, err := bs.RefFromHex(blobInfo.Name())
				if err != nil {
					continue
				}

				err = f(ref)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ListAnchors lists all anchors in the store, in lexical order.
func (s *Store) ListAnchors(ctx context.Context, start bs.Anchor, f func(bs.Anchor) error) error {
	topLevel, err := ioutil.ReadDir(s.anchorroot())
	if err != nil {
		return errors.Wrapf(err, "reading dir %s", s.anchorroot())
	}
	// xxx filter results

	var anchors []bs.Anchor

	for _, topInfo := range topLevel {
		var (
			topName = topInfo.Name()
			topDir  = filepath.Join(s.anchorroot(), topName)
		)
		midLevel, err := ioutil.ReadDir(topDir)
		if err != nil {
			return errors.Wrapf(err, "reading dir %s", topDir)
		}
		// xxx filter results

		for _, midInfo := range midLevel {
			var (
				midName = midInfo.Name()
				midDir  = filepath.Join(topDir, midName)
			)
			anchorLevel, err := ioutil.ReadDir(midDir)
			if err != nil {
				return errors.Wrapf(err, "reading dir %s", midDir)
			}
			// xxx filter results

			for _, anchorInfo := range anchorLevel {
				if err = ctx.Err(); err != nil {
					return err
				}
				anchor, err := decodeAnchor(anchorInfo.Name())
				if err != nil {
					return errors.Wrapf(err, "decoding anchor %s", anchorInfo.Name())
				}
				if anchor <= start {
					continue
				}
				anchors = append(anchors, anchor)
			}
		}
	}

	sort.Slice(anchors, func(i, j int) bool { return anchors[i] < anchors[j] })

	for _, anchor := range anchors {
		err = f(anchor)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListAnchorRefs lists all blob refs for a given anchor,
// together with their timestamps,
// in chronological order.
func (s *Store) ListAnchorRefs(ctx context.Context, a bs.Anchor, f func(bs.TimeRef) error) error {
	path := s.anchorpath(a)
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrapf(err, "reading dir %s", path)
	}
	// xxx filter entries and sort by parsed time
	for _, entry := range entries {
		name := entry.Name()
		t, err := time.Parse(time.RFC3339Nano, name)
		if err != nil {
			return errors.Wrapf(err, "parsing time from %s", name)
		}
		h, err := ioutil.ReadFile(filepath.Join(path, name))
		if err != nil {
			return errors.Wrapf(err, "reading file %s/%s", path, name)
		}
		ref, err := bs.RefFromHex(string(h))
		if err != nil {
			return errors.Wrapf(err, "hex-decoding %s", string(h))
		}
		err = f(bs.TimeRef{T: t, R: ref})
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeAnchor(a bs.Anchor) string {
	return url.PathEscape(string(a))
}

func decodeAnchor(inp string) (bs.Anchor, error) {
	out, err := url.PathUnescape(inp)
	return bs.Anchor(out), errors.Wrapf(err, "unescaping anchor %s", inp)
}

func init() {
	store.Register("file", func(_ context.Context, conf map[string]interface{}) (bs.Store, error) {
		root, ok := conf["root"].(string)
		if !ok {
			return nil, errors.New(`missing "root" parameter`)
		}
		return New(root), nil
	})
}
