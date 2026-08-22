$ErrorActionPreference = "Stop"

function Invoke-NativeChecked([string]$StepName, [scriptblock]$Script) {
    Write-Host "::group::$StepName"
    Write-Host "[CI-STEP] Starting: $StepName" -ForegroundColor Cyan
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        & $Script
        $exit = $LASTEXITCODE
        $sw.Stop()
        if ($exit -ne 0) {
            Write-Host "[CI-STEP-FAIL] $StepName failed with exit code: $exit" -ForegroundColor Red
            Write-Host "::endgroup::"
            exit $exit
        }
        Write-Host "[CI-STEP-PASS] $StepName completed successfully in $($sw.Elapsed.TotalSeconds.ToString('F2'))s" -ForegroundColor Green
        Write-Host "::endgroup::"
    } catch {
        $sw.Stop()
        Write-Host "[CI-STEP-ERROR] $StepName encountered exception: $_" -ForegroundColor Red
        Write-Host "::endgroup::"
        exit 1
    }
}

$env:GOROOT = "C:\DevTools\go"
$env:GOPATH = "C:\DevTools\go-path"
$env:Path = "C:\DevTools\go\bin;C:\DevTools\go-path\bin;C:\DevTools\node;" + $env:Path

Invoke-NativeChecked "Diagnostics" {
    go version
    node --version
}

Invoke-NativeChecked "Validate Frontend JS" {
    node --check frontend/dist/app.js
}

Invoke-NativeChecked "Download Dependencies" {
    go mod download
}

Invoke-NativeChecked "Test Suite" {
    go test -cover .
}

Invoke-NativeChecked "Go Vet" {
    go vet .
}

Invoke-NativeChecked "Build Desktop App" {
    wails build -clean
}

Invoke-NativeChecked "Verify Executable" {
    if (-not (Test-Path "build/bin/CsvToJson.exe")) {
        throw "CsvToJson.exe was not produced"
    }
    Get-Item "build/bin/CsvToJson.exe" | Format-List Name,Length,LastWriteTime
    Get-FileHash "build/bin/CsvToJson.exe" -Algorithm SHA256
}

Write-Host "`n[CI-SUCCESS] All csv-to-json Windows CI steps completed successfully!" -ForegroundColor Green
