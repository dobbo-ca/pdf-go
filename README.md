# pdf-go

pdf-go is a dobbo-ca-owned hard fork of [pdfcpu](https://github.com/pdfcpu/pdfcpu), kept for
kleio and byblos. It carries per-parse resource limits so a hostile PDF cannot exhaust memory during a parse.
We pull upstream changes only when we want them, and nothing is sent upstream. Upstream's Apache-2.0
LICENSE is kept (upstream has no NOTICE file).

Import `github.com/dobbo-ca/pdf-go` directly; no `replace` directive is needed.

The patches add two per-parse limits to `model.Configuration.Limits`. Zero keeps pdfcpu's behaviour (no entry cap, and for decoded bytes the separate defaults below). When a limit is hit the
error satisfies `errors.Is(err, model.ErrLimitExceeded)`; for the decode budgets that wrap is the only change to the error at zero:

- `MaxObjectStreamTotalCount`: object stream entries (`/N`) summed over all object streams of one parse.
- `MaxParsedBytes`: estimated heap of the array elements and dictionary entries parsed over one parse. A few bytes of `[]` parse into tens of bytes of heap, which the decoded-byte limits do not see.
- `MaxTotalDecodeBytes`: decoded bytes charged over one parse, counted as buffer capacity like pdfcpu's own budgets. When set, one budget covers content, xref, font and metadata
  streams, object streams and images, and decoding a stream again charges it again. Zero keeps pdfcpu's behaviour:
  `DecodeStream` calls and object streams each have their own default allowance derived from `MaxDecodeBytes`, and images have none.

## Versions

The first release under `github.com/dobbo-ca/pdf-go` is v0.17.0, not v0.1.0. This repo once held pdfcpu's
v0.0.1-v0.16.4 tags, and proxy.golang.org resolved them under this module path (for example v0.1.1 and
v0.16.4), so those version numbers may be cached and are never reused.

The upstream README follows.

---

# pdfcpu: PDF tooling for Go and the command line

[![Go Reference](https://pkg.go.dev/badge/github.com/dobbo-ca/pdf-go.svg)](https://pkg.go.dev/github.com/dobbo-ca/pdf-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="left">
  <a href="https://pdfcpu.io"><img src="resources/logoSmall.png" width="150"></a>
</p>

pdfcpu is a PDF processing library and command-line tool written in Go. It supports validation, optimization,
encryption, signing, document assembly, content extraction, and other common PDF operations.

pdfcpu supports PDF versions through PDF 2.0 (ISO 32000-2). PDF 2.0 validation support is basic and continuously
improving.

<p align="left">
  <a href="https://pdfa.org"><img src="resources/pdfa.png" width="75" align="middle"></a>
  &nbsp; Horst Rutter, the maintainer of pdfcpu, is a member of the PDF Association.
</p>

---

## Installation

```
go get github.com/dobbo-ca/pdf-go
```


---

## Usage

### CLI

Validate a PDF:
```
pdfcpu validate input.pdf
```

Merge two PDFs:
```
pdfcpu merge merged.pdf in1.pdf in2.pdf
```

### Go API

See API documentation for usage examples.

### Local images in create and form input

Local images in `create`, `form fill`, and JSON/CSV `form multifill` are confined to
an allowed directory. <br>
API callers must set `conf.LocalImageRoot` to enable them. The CLI defaults to its working
directory and `localImageRoot` in `config.yml` selects another root.

See [local image access](https://pdfcpu.io/create/create/#local-image-access) for configuration examples, path
restrictions, migration instructions and security limitations.

---

## Features

* Validate, optimize, split, trim, and merge PDFs
* Extract and manipulate images, fonts, and metadata
* Encrypt and decrypt PDFs
* Resize and rotate pages
* Add and remove stamps and watermarks
* Validate signature integrity, report signature evidence, and remove signatures
* Manage attachments and portfolios

## Examples

Selected examples:

<p align="center">
  <kbd><a href="https://pdfcpu.io/generate/grid"><img src="resources/gridpdf.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/core/watermark"><img src="resources/wmi1abs.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/generate/nup"><img src="resources/nup9pdf.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/fonts/fonts"><img src="resources/cjkv.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/core/stamp"><img src="resources/4exp.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/form/form"><img src="resources/form.png" height="150"></a></kbd><br><br>
  <kbd><a href="https://pdfcpu.io/create/create"><img src="resources/table.png" height="100"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/core/stamp"><img src="resources/sti.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/core/watermark"><img src="resources/wmi4.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/create/create"><img src="resources/imagebox.png" height="100"></a></kbd>&nbsp;<br><br>
  <kbd><a href="https://pdfcpu.io/generate/booklet"><img src="resources/book2A4p1.png" height="150"></a></kbd>
  <kbd><a href="https://pdfcpu.io/core/stamp"><img src="resources/stp.png" height="150"></a></kbd>&nbsp;
  <kbd><a href="https://pdfcpu.io/generate/grid"><img src="resources/gridimg.png" height="150"></a></kbd>
  <kbd><a href="https://pdfcpu.io/core/stamp"><img src="resources/stRoundBorder.png" height="150"></a></kbd>
  <kbd><a href="https://pdfcpu.io/generate/booklet"><img src="resources/book4A4p1.png" height="150"></a></kbd>
</p>

---

## Commands

Complete list of supported commands:

| | | |
|---|---|---|
| [annotations](https://pdfcpu.io/annot/annot) | [attachments](https://pdfcpu.io/attach/attach) | [booklet](https://pdfcpu.io/generate/booklet) |
| [bookmarks](https://pdfcpu.io/bookmarks/bookmarks) | [boxes](https://pdfcpu.io/boxes/boxes) | [certificates](https://pdfcpu.io/core/certs) |
| [change owner password](https://pdfcpu.io/encrypt/change_opw) | [change user password](https://pdfcpu.io/encrypt/change_upw) | [collect](https://pdfcpu.io/core/collect) |
| [config](https://pdfcpu.io/config/config) | [create](https://pdfcpu.io/create/create) | [crop](https://pdfcpu.io/core/crop) |
| [cut](https://pdfcpu.io/generate/cut) | [decrypt](https://pdfcpu.io/encrypt/decryptpdf) | [encrypt](https://pdfcpu.io/encrypt/encryptpdf) |
| [extract](https://pdfcpu.io/extract/extract) | [fonts](https://pdfcpu.io/fonts/fonts) | [form](https://pdfcpu.io/form/form) |
| [grid](https://pdfcpu.io/generate/grid) | [images](https://pdfcpu.io/images/images) | [import](https://pdfcpu.io/generate/import) |
| [info](https://pdfcpu.io/info) | [keywords](https://pdfcpu.io/keywords/keywords) | [merge](https://pdfcpu.io/core/merge) |
| [ndown](https://pdfcpu.io/generate/ndown) | [nup](https://pdfcpu.io/generate/nup) | [optimize](https://pdfcpu.io/core/optimize) |
| [pagelayout](https://pdfcpu.io/pagelayout/pagelayout) | [pagemode](https://pdfcpu.io/pagemode/pagemode) | [pages](https://pdfcpu.io/pages/pages) |
| [permissions](https://pdfcpu.io/encrypt/perm_set) | [portfolio](https://pdfcpu.io/portfolio/portfolio) | [poster](https://pdfcpu.io/generate/poster) |
| [properties](https://pdfcpu.io/properties/properties) | [resize](https://pdfcpu.io/core/resize) | [rotate](https://pdfcpu.io/core/rotate) |
| [signatures](https://pdfcpu.io/core/sign) | [split](https://pdfcpu.io/core/split) | [stamp](https://pdfcpu.io/core/stamp) |
| [trim](https://pdfcpu.io/core/trim) | [validate](https://pdfcpu.io/core/validate) | [viewerpref](https://pdfcpu.io/viewerpref/viewerpref) |
| [watermark](https://pdfcpu.io/core/watermark) | [zoom](https://pdfcpu.io/core/zoom) | |

---

## Motivation

pdfcpu aims to provide comprehensive PDF processing capabilities implemented in Go for both individual files and
automated batch processing.

It focuses on correctness, robustness and independence from external dependencies.

---

## Focus

- comprehensive PDF processing functionality  
- minimal external dependencies  
- predictable and stable behavior  

---

## Documentation

* Project documentation: https://pdfcpu.io
* [Security policy](SECURITY.md)

### CLI

* Command help: `pdfcpu [command] --help`

### Go API

* Package documentation: https://pkg.go.dev/github.com/dobbo-ca/pdf-go
* API documentation: https://pkg.go.dev/github.com/dobbo-ca/pdf-go/pkg/api
* Examples:

  * https://github.com/dobbo-ca/pdf-go/tree/main/pkg/api/test
  * https://github.com/dobbo-ca/pdf-go/tree/main/pkg/samples

---

## Contributing

Contributions are welcome.

### Reporting PDF issues

For triage, we use Adobe Acrobat Reader and macOS Preview as practical compatibility references.
<br>
Reports are especially helpful when a PDF opens in either application but cannot be processed by pdfcpu, as these cases may
reveal opportunities to improve validation or parser compatibility.
<br>
If neither application can open the PDF, it is unlikely that pdfcpu will be able to process it reliably.

Start by validating the file:

```bash
pdfcpu validate -vv <file.pdf>
```

Include the command, its verbose output, and a sample PDF with the report.
<br>
Please submit only files that you have permission to share and that contain no confidential information or personal data.

---

## Security

Please do not report security vulnerabilities through public GitHub issues.
<br>
See the [security policy](SECURITY.md) for private reporting instructions.

---

## Disclaimer

Use of pdfcpu assumes compliance with all applicable copyrights for any processed PDF content, including embedded resources such as fonts and images.

Gopher artwork by [Renee French](https://www.instagram.com/reneefrenchphoto)

---

## License

Apache License 2.0

---

<img referrerpolicy="no-referrer-when-downgrade" src="https://static.scarf.sh/a.png?x-pxid=592f747a-ac99-42e3-802a-80dbf2b94519" width="1" height="1" />
