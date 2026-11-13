package pdfcpu

import (
	"context"

	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/model"
	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/types"
)

type lazyImageKey struct{}

// WithLazyImageStreams returns a context that makes Read leave the encoded bytes of image XObject streams
// unloaded (sd.Raw nil) so a read-only caller holds one page's images at a time instead of the whole file's.
// The caller loads an image's bytes with LoadStreamRaw and must keep the reader open while it may do so.
// Unencrypted files only: an encrypted file loads eagerly.
func WithLazyImageStreams(c context.Context) context.Context {
	return context.WithValue(c, lazyImageKey{}, true)
}

// LoadStreamRaw loads sd.Raw of a stream left unloaded by WithLazyImageStreams. It does nothing if Raw is loaded.
// sd may be a copy of the table's entry, so the bytes live only as long as the copy.
func LoadStreamRaw(c context.Context, ctx *model.Context, sd *types.StreamDict) error {
	return loadEncodedStreamContentCharged(c, ctx, sd, false, false)
}

// deferImageStream charges sd's declared length and reports true when it leaves an image stream's bytes unread.
func deferImageStream(c context.Context, ctx *model.Context, sd *types.StreamDict, fixLength bool) (bool, error) {
	if lazy, _ := c.Value(lazyImageKey{}).(bool); !lazy || fixLength || ctx.EncKey != nil {
		return false, nil
	}
	image, err := isImageStreamDict(ctx, sd)
	if err != nil || !image {
		return false, err
	}
	if err := ensureIndirectStreamLength(c, ctx, sd, false); err != nil {
		return false, err
	}
	if sd.StreamLength == nil || *sd.StreamLength <= 0 { // unknown length: only a read can tell
		return false, nil
	}
	n := min(*sd.StreamLength, max(ctx.Read.FileSize-sd.StreamOffset, 0))
	if err := chargeObjectRead(c, ctx, n); err != nil {
		return false, err
	}
	return true, ctx.XRefTable.ChargeStreamBytes(n)
}
