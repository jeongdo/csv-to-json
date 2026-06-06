# CsvToJson

A lightweight desktop CSV to JSON converter built with Go.

CsvToJson is a portable desktop utility that converts CSV files into JSON format with a simple graphical interface.

The application runs entirely on the local machine, requires no installation, and has no external dependencies.

---

## Features

* Single executable deployment
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
* Automatic process cleanup on application exit

---

## How It Works

```text
CSV File
    ↓
CsvToJson
    ↓
CSV Validation
    ↓
Streaming Conversion
    ↓
JSON Download
```

The application starts a temporary local HTTP server, launches a browser window in App Mode, converts CSV records into JSON format, and automatically downloads the generated file.

All processing occurs locally on the user's machine.

---

## Supported Data Types

The converter automatically detects common JSON data types.

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

Supported types:

* String
* Integer
* Float
* Boolean

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

Values with meaningful leading zeros remain strings to prevent accidental corruption of postal codes, employee IDs, product codes, and similar identifiers.

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
CsvToJson
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

```bash
go build -ldflags="-H windowsgui" -o CsvToJson.exe
```

---

## Requirements

* Windows 10+
* Go 1.24+ (development only)

No installation is required for end users.

---

## Limitations

Current version intentionally focuses on simplicity.

Not currently supported:

* XLSX input
* Nested JSON generation
* Schema mapping
* Custom output formatting
* Automatic delimiter detection

---

## License

MIT License

Feel free to use, modify, and distribute.
