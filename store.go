package bs

import (
	"context"
	"errors"
	"time"
)

// Getter is a read-only Store (qv).
type Getter interface {
	// Get gets a blob by its ref.
	Get(context.Context, Ref) (Blob, error)

	// GetMulti returns multiple blobs in a single call.
	GetMulti(context.Context, []Ref) (GetMultiResult, error)

	// GetAnchor gets the latest ref with the given anchor
	// not later than the given timestamp.
	GetAnchor(context.Context, Anchor, time.Time) (Ref, error)

	// ListRefs lists all blob refs in the store in lexical order,
	// beginning with the first ref _after_ the specified one.
	//
	// The contents of the channel must reflect at least the set of refs known at the moment ListRefs was called.
	// It is unspecified whether later changes to that set are reflected in the channel while it is being consumed.
	//
	// On exit ListRefs must close the channel.
	ListRefs(context.Context, Ref, chan<- Ref) error

	// ListAnchors lists all anchors in the store in lexical order,
	// beginning with the first anchor _after_ the specified one.
	//
	// The contents of the channel must reflect at least the set of anchors known at the moment ListAnchors was called.
	// It is unspecified whether later changes to that set are reflected in the channel while it is being consumed.
	//
	// On exit ListAnchors must close the channel.
	ListAnchors(context.Context, Anchor, chan<- Anchor) error

	// ListAnchorRefs lists all TimeRefs for the given anchor,
	// in time order.
	//
	// The contents of the channel must reflect at least the set of time/ref pairs known at the moment ListAnchorRefs was called.
	// It is unspecified whether later changes to that set are reflected in the channel while it is being consumed.
	//
	// On exit ListAnchorRefs must close the channel.
	ListAnchorRefs(context.Context, Anchor, chan<- TimeRef) error
}

// Store is a blob store.
// It stores byte sequences - "blobs" - of arbitrary length.
// Each blob can be retrieved using its "ref" as a lookup key.
// A ref is simply the SHA2-256 hash of the blob's content.
type Store interface {
	Getter

	// Put adds a blob to the store if it was not already present.
	// It returns the blob's ref and a boolean that is true iff the blob had to be added.
	Put(context.Context, Blob) (ref Ref, added bool, err error)

	// PutMulti adds multiple blobs to the store.
	PutMulti(context.Context, []Blob) (PutMultiResult, error)

	// PutAnchor associates an anchor and a timestamp with a ref.
	PutAnchor(context.Context, Ref, Anchor, time.Time) error
}

type (
	// GetMultiResult is the type of value returned by Getter.GetMulti.
	GetMultiResult map[Ref]func(context.Context) (Blob, error)

	// PutMultiResult is the type of value returned by Store.PutMulti.
	PutMultiResult []func(context.Context) (Ref, bool, error)
)

// ErrNotFound is the error returned
// when a Getter tries to access a non-existent ref or anchor.
var ErrNotFound = errors.New("not found")
