package pdfcpu

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/model"
	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/types"
)

// lazyImagePDF builds pages one-image pages; image i is n bytes of i-dependent noise.
func lazyImagePDF(pages, n int) ([]byte, [][]byte) {
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
		fmt.Fprintf(&b, "%d 0 obj\n<</Type/XObject/Subtype/Image/Width 1/Height 1/ColorSpace/DeviceGray/BitsPerComponent 8/Filter/DCTDecode/Length %d>>\nstream\n", len(offs), n)
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
