# CSV to JSON

A lightweight, local-first Windows desktop converter built with Go and Wails.

CSV to JSON converts CSV/TSV/delimited text files to JSON without uploading data anywhere. The desktop UI passes only a file path to Go; the converter opens the source directly and processes records incrementally, so large files do not need to be loaded into the browser or held entirely in memory.

## Highlights

- Native Wails desktop window — no localhost server and no Chrome/Edge app-mode process
- Drag & drop and native Open/Save dialogs
- Direct disk streaming with `encoding/csv`
- Bounded preview: header + first 8 rows only
- Automatic delimiter detection: comma, tab, pipe, semicolon
- Valid single-column CSV support
- UTF-8 BOM removal from headers
- Empty/duplicate header validation
- Header order preserved in JSON output
- Optional native JSON type inference
- Optional empty-cell → `null`
- Leading-zero identifiers preserved as strings
- Arbitrarily large valid JSON numbers emitted without `float64` precision loss
- Original string whitespace preserved
- Progress reporting for large files
- Safe output commit: write to a temporary file first, then replace the target only after the full CSV succeeds
- Korean / English UI
- Light / dark theme
- Local-only processing

## Data safety

The conversion path is intentionally file-based:

```text
CSV / TSV file
    │
    ├─ Inspect: header + up to 8 preview rows
    │
    └─ Convert
         │
         ▼
    Go opens source file
         │
         ▼
    csv.Reader (record-by-record)
         │
         ▼
    temporary JSON file
         │
         ├─ error → delete temporary file, keep existing output untouched
         │
         └─ success → replace target JSON
```

The frontend never calls `File.arrayBuffer()`, `File.text()`, `response.blob()`, or a browser upload endpoint for conversion.

## Type inference

With **Infer data types** enabled:

```csv
name,age,zipcode,active,big
Kim,20,01001,true,9223372036854775808
```

becomes:

```json
[
  {
    "name": "Kim",
    "age": 20,
    "zipcode": "01001",
    "active": true,
    "big": 9223372036854775808
  }
]
```

Non-JSON numeric spellings such as `+1`, `.5`, `1.` and `-01` remain strings instead of producing invalid JSON.

## Project structure

```text
.
├─ main.go                  # Wails application entry point
├─ app.go                   # Native dialogs / UI bindings
├─ converter.go             # Pure CSV → JSON conversion core
├─ file_io.go               # Direct file streaming / atomic output
├─ converter_test.go
├─ file_io_test.go
├─ frontend/
│  └─ dist/
│     ├─ index.html
│     ├─ app.js
│     └─ style.css
├─ build/
│  └─ windows/
│     └─ icon.ico
├─ .github/workflows/
│  └─ windows-build.yml
└─ wails.json
```

## Build

Development requirements:

- Go 1.26+
- Wails v2.13.0
- Windows environment for the Windows executable

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
go test .
wails build -clean
```

The Windows executable is produced at:

```text
build/bin/CsvToJson.exe
```

The repository also contains a GitHub Actions workflow that runs the Go tests and builds the Windows executable on every push to `main`.

## Current scope

Intentionally not included yet:

- XLSX input
- Nested JSON mapping
- Schema mapping/editor
- Arbitrary output templates

The goal is to keep one job — reliable CSV → JSON conversion — small, fast and trustworthy.

## License

Apache License 2.0. See [LICENSE](LICENSE).
