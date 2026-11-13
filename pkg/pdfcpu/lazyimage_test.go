package pdfcpu

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"runtime"
	"testing"

	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/model"
	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/types"
)

// lazyImagePDF builds pages one-image pages; image i is n bytes of i-dependent noise.
func lazyImagePDF(pages, n int) ([]byte, [][]byte) {
	return lazyImagePDFType(pages, n, "/Type/XObject")
}

// lazyImagePDFType is lazyImagePDF with the image dict's /Type entry set to typ (empty: no /Type).
func lazyImagePDFType(pages, n int, typ string) ([]byte, [][]byte) {
	var b bytes.Buffer
	var offs []int
	obj := func(body string) {
		offs = append(offs, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", len(offs), body)
	}
	b.WriteString("%PDF-1.4\n")
	kids := ""
	for i := 0; i < pages; i++ {
		kids += fmt.Sprintf("%d 0 R ", 3+i*3)
	}
	obj("<</Type/Catalog/Pages 2 0 R>>")
	obj(fmt.Sprintf("<</Type/Pages/Count %d/Kids[%s]>>", pages, kids))
	imgs := make([][]byte, pages)
	for i := 0; i < pages; i++ {
		base := 3 + i*3
		obj(fmt.Sprintf("<</Type/Page/Parent 2 0 R/MediaBox[0 0 10 10]/Contents %d 0 R/Resources<</XObject<</I %d 0 R>>>>>>", base+1, base+2))
		c := "q 10 0 0 10 0 0 cm /I Do Q"
		obj(fmt.Sprintf("<</Length %d>>\nstream\n%s\nendstream", len(c), c))
		img := make([]byte, n)
		for j := range img {
			img[j] = byte(j*31 + i*7 + j>>8)
		}
		imgs[i] = img
		offs = append(offs, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n<<%s/Subtype/Image/Width 1/Height 1/ColorSpace/DeviceGray/BitsPerComponent 8/Filter/DCTDecode/Length %d>>\nstream\n", len(offs), typ, n)
		b.Write(img)
		b.WriteString("\nendstream\nendobj\n")
	}
	x := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offs)+1)
	for _, o := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<</Size %d/Root 1 0 R>>\nstartxref\n%d\n%%%%EOF\n", len(offs)+1, x)
	return b.Bytes(), imgs
}

func readLazyImagePDF(t *testing.T, pdf []byte, lazy bool) *model.Context {
	t.Helper()
	c := context.Background()
	if lazy {
		c = WithLazyImageStreams(c)
	}
	ctx, err := Read(c, bytes.NewReader(pdf), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func heapAfterGC() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// Image bytes are not retained by a lazy read, load on demand, and equal the eager bytes.
func TestLazyImageStreams(t *testing.T) {
	const pages, n = 24, 512 << 10
	pdf, imgs := lazyImagePDF(pages, n)
	total := uint64(pages * n)

	before := heapAfterGC()
	ctx := readLazyImagePDF(t, pdf, true)
	grew := heapAfterGC() - before
	if grew > total/10 {
		t.Errorf("lazy read retains %d bytes of %d stream bytes", grew, total)
	}

	for i := 0; i < pages; i++ {
		sd := ctx.XRefTable.Table[5+i*3].Object.(types.StreamDict)
		if sd.Raw != nil {
			t.Fatalf("image %d: Raw loaded by lazy read", i)
		}
		if err := LoadStreamRaw(context.Background(), ctx, &sd); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(sd.Raw, imgs[i]) {
			t.Fatalf("image %d: lazy bytes differ", i)
		}
	}

	eager := readLazyImagePDF(t, pdf, false)
	if sd := eager.XRefTable.Table[5].Object.(types.StreamDict); !bytes.Equal(sd.Raw, imgs[0]) {
		t.Fatal("eager read lost the image bytes")
	}
}

// /Type is optional on an image XObject: a stream with only /Subtype /Image still loads lazily.
func TestLazyImageStreamsNoType(t *testing.T) {
	pdf, imgs := lazyImagePDFType(2, 4<<10, "")
	ctx := readLazyImagePDF(t, pdf, true)
	for i := range imgs {
		sd := ctx.XRefTable.Table[5+i*3].Object.(types.StreamDict)
		if sd.Raw != nil {
			t.Fatalf("image %d without /Type: Raw loaded by lazy read", i)
		}
		if err := LoadStreamRaw(context.Background(), ctx, &sd); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(sd.Raw, imgs[i]) {
			t.Fatalf("image %d: lazy bytes differ", i)
		}
	}
}

func zlibBytes(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
}

// A Flate image with a Flate /SMask renders the same lazy as eager: the mask is a deferred image stream too.
func TestLazyImageSoftMask(t *testing.T) {
	const w, h = 8, 4
	pix, mask := make([]byte, w*h), make([]byte, w*h)
	for i := range pix {
		pix[i], mask[i] = byte(i*9), byte(255-i*5)
	}
	img, sm := zlibBytes(t, pix), zlibBytes(t, mask)

	var b bytes.Buffer
	var offs []int
	obj := func(body string, stream []byte) {
		offs = append(offs, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s", len(offs), body)
		if stream != nil {
			b.WriteString("\nstream\n")
			b.Write(stream)
			b.WriteString("\nendstream")
		}
		b.WriteString("\nendobj\n")
	}
	b.WriteString("%PDF-1.4\n")
	obj("<</Type/Catalog/Pages 2 0 R>>", nil)
	obj("<</Type/Pages/Count 1/Kids[3 0 R]>>", nil)
	obj("<</Type/Page/Parent 2 0 R/MediaBox[0 0 10 10]/Resources<</XObject<</I 4 0 R>>>>>>", nil)
	g := fmt.Sprintf("/Type/XObject/Subtype/Image/Width %d/Height %d/ColorSpace/DeviceGray/BitsPerComponent 8/Filter/FlateDecode", w, h)
	obj(fmt.Sprintf("<<%s/SMask 5 0 R/Length %d>>", g, len(img)), img)
	obj(fmt.Sprintf("<<%s/Length %d>>", g, len(sm)), sm)
	x := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offs)+1)
	for _, o := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<</Size %d/Root 1 0 R>>\nstartxref\n%d\n%%%%EOF\n", len(offs)+1, x)

	render := func(lazy bool) []byte {
		ctx := readLazyImagePDF(t, b.Bytes(), lazy)
		sd := ctx.XRefTable.Table[4].Object.(types.StreamDict)
		if err := LoadStreamRaw(context.Background(), ctx, &sd); err != nil {
			t.Fatal(err)
		}
		im, err := ExtractImage(context.Background(), ctx, &sd, false, "I", 4, false)
		if err != nil {
			t.Fatalf("lazy=%v: %v", lazy, err)
		}
		out, _ := io.ReadAll(im)
		return out
	}
	if want, got := render(false), render(true); !bytes.Equal(want, got) {
		t.Fatal("lazy render differs from eager")
	}
}
