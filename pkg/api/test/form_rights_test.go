/*
Copyright 2026 The pdf Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/dobbo-ca/pdf-go/pkg/api"
	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/model"
	"github.com/dobbo-ca/pdf-go/pkg/pdfcpu/types"
)

// TestFillFormDropsXFAAndPerms checks that a fill removes /XFA and /Perms even
// when the form is not signed (no SigFlags). A full rewrite breaks a UR3
// usage-rights signature, and a stale XFA stream overrides the filled
// AcroForm values in XFA-aware viewers.
func TestFillFormDropsXFAAndPerms(t *testing.T) {
	in := filepath.Join(samplesDir, "form", "demoSinglePage", "english.pdf")
	jsonIn := filepath.Join(samplesDir, "form", "fill", "english.json")

	// Build the fixture: the sample plus an XFA stream and a /Perms dict.
	f, err := os.Open(in)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	conf := model.NewDefaultConfiguration()
	ctx, err := api.ReadValidateAndOptimize(t.Context(), f, conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Form == nil {
		t.Fatal("fixture has no AcroForm")
	}
	if _, ok := ctx.Form["SigFlags"]; ok {
		t.Fatal("fixture must be unsigned")
	}
	xfa, err := ctx.NewStreamDictForBuf([]byte("<xdp:xdp/>"))
	if err != nil {
		t.Fatal(err)
	}
	if err := xfa.Encode(); err != nil {
		t.Fatal(err)
	}
	xfaRef, err := ctx.IndRefForNewObject(*xfa)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Form["XFA"] = *xfaRef
	ctx.RootDict["AcroForm"] = ctx.Form
	ctx.RootDict["Perms"] = types.Dict{"UR3": types.Dict{"Type": types.Name("Sig")}}
	var fixture bytes.Buffer
	if err := api.Write(t.Context(), ctx, &fixture, conf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(fixture.Bytes(), []byte("/XFA")) || !bytes.Contains(fixture.Bytes(), []byte("/Perms")) {
		t.Fatal("fixture lost /XFA or /Perms before the fill")
	}

	data, err := os.ReadFile(jsonIn)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := api.FillForm(t.Context(), bytes.NewReader(fixture.Bytes()), bytes.NewReader(data), &out, model.NewDefaultConfiguration()); err != nil {
		t.Fatal(err)
	}

	got, err := api.ReadValidateAndOptimize(t.Context(), bytes.NewReader(out.Bytes()), model.NewDefaultConfiguration(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.RootDict["Perms"]; ok {
		t.Error("filled output still has /Perms")
	}
	if got.Form == nil {
		t.Fatal("filled output lost its AcroForm")
	}
	if _, ok := got.Form["XFA"]; ok {
		t.Error("filled output still has /XFA")
	}
}
