# csv-to-json
[🚀 Download Latest CsvToJson.exe](https://github.com/jeongdo/csv-to-json/releases/latest)

A lightweight desktop CSV to JSON converter built with Go.

csv-to-json is a portable desktop utility that converts CSV files into JSON format through a simple graphical interface.

The application runs entirely on the local machine, requires no installation, and has no external dependencies.

Download the executable, run it, and convert your CSV files to JSON instantly.

---

## Features

* Single executable deployment
* Automatic delimiter detection (CSV, TSV, pipe, semicolon)
* Embedded HTML, CSS, JavaScript and Icon assets using Go Embed
* Local-only processing (no external network communication)
* Automatic JSON file download
* CSV header validation
* Duplicate header detection
* Empty header detection
* UTF-8 BOM removal
* Original CSV column order preservation
* Automatic JSON type inference
* Leading zero protection
* Streaming CSV conversion for reduced memory usage
* Korean / English UI localization
* Chrome / Edge App Mode desktop UI
* Toast notification based error handling
* Drag & Drop CSV file support
* Automatic process cleanup on application exit

---

## How It Works

```text
CSV File
    ↓
csv-to-json
    ↓
CSV Validation
    ↓
Streaming Conversion
    ↓
JSON Download
```

The application starts a temporary local HTTP server, launches a browser window in App Mode, converts CSV records into
JSON format, and automatically downloads the generated file.

All processing occurs locally on the user's machine.

---

## Supported Data Types

csv-to-json automatically analyzes input data and applies the optimal JSON data types.

Supported types:

* String
* Integer
* Float
* Boolean

### Input CSV

```csv
name,age,salary,active
Kim,20,3500.50,true
Lee,30,4200,false
```

### Output JSON

```json
[
  {
    "name": "Kim",
    "age": 20,
    "salary": 3500.50,
    "active": true
  },
  {
    "name": "Lee",
    "age": 30,
    "salary": 4200,
    "active": false
  }
]
```

---

## Leading Zero Protection

### Input

```csv
zipcode
01001
```

### Output

```json
{
  "zipcode": "01001"
}
```

Values with meaningful leading zeros remain strings to prevent accidental corruption of postal codes, employee IDs,
product codes, and similar identifiers.

---

## CSV Validation

Validation checks:

* Empty headers
* Duplicate headers
* Missing header row
* CSV parsing errors

### Invalid Example

```csv
name,name,age
Kim,20,Seoul
```

### Result

```text
Duplicate header detected: name
```

---

## Architecture

```text
CsvToJson.exe
        │
        ▼
Embedded HTTP Server
(127.0.0.1 Random Port)
        │
        ▼
Chrome / Edge App Mode
        │
        ▼
HTML + CSS + JavaScript UI
        │
        ▼
Streaming CSV Parser
        │
        ▼
JSON Download
```

---

## Browser Support

Windows launch priority:

1. Google Chrome
2. Microsoft Edge
3. Default Browser

If Chrome or Edge is unavailable, the application automatically falls back to the system default browser.

---

## Project Structure

```text
csv-to-json
│
├── main.go
├── browser.go
├── converter.go
├── index.html
├── style.css
├── app.js
├── app.ico
└── README.md
```

---

## Build

### 1. Prerequisites (Required Tools Installation)

Install `goversioninfo` for handling icons and Windows resource embedding.

```bash
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

2. Resource Files (Optional)

To include an icon and version information in the executable, place the following files in the project root:

* app.ico : Application icon
* versioninfo.json : Version information configuration file
* app.manifest : Windows application manifest (execution permissions)

### 3. Generate Resource File (.syso)

```bash
goversioninfo -platform-specific=true
```

### 4. Final Build

```bash
go build -ldflags="-H windowsgui" -o CsvToJson.exe
```

---

## Requirements

* Windows 10+
* Go 1.26.4 (development only)

No installation is required for end users.

---

## Limitations

Current version intentionally focuses on simplicity.

Not supported:

* XLSX input
* Nested JSON generation
* Schema mapping
* Custom output formatting

---

## License

This project is licensed under the Apache License, Version 2.0.  
See the `LICENSE` file for details.

---

## Contributing

All contributions to this project are subject to the Apache License, Version 2.0.  
By submitting a contribution, you agree that your code will be distributed under the project's license and managed by the original author.