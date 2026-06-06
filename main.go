// main.go
package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

//go:embed index.html
var htmlContent string

//go:embed style.css
var cssContent string

//go:embed app.js
var jsContent string

//go:embed app.ico
var iconBytes []byte

var server *http.Server

func killExistingProcesses() {
	currentPID := os.Getpid()

	if runtime.GOOS == "windows" {
		// 실행 파일의 이름을 동적으로 가져옴 (예: C:\Tool\CsvToJson.exe -> CsvToJson.exe)
		exePath, err := os.Executable()
		if err != nil {
			return
		}
		exeName := filepath.Base(exePath)

		cmd := exec.Command(
			"taskkill",
			"/F",
			"/IM",
			exeName, // 하드코딩 대신 동적 변수 사용
			"/FI",
			fmt.Sprintf("PID ne %d", currentPID),
		)

		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}

		cmd.Run()
		// 200ms 정도는 여유를 주는 게 시스템상 안정적입니다.
		time.Sleep(200 * time.Millisecond)
	}
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	go func() {
		time.Sleep(500 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		server.Shutdown(ctx)
		os.Exit(0)
	}()
}

func main() {
	go killExistingProcesses()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/convert", convertHandler)
	mux.HandleFunc("/shutdown", shutdownHandler)

	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Content-Type",
			"application/javascript; charset=utf-8",
		)
		io.WriteString(w, jsContent)
	})

	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		io.WriteString(w, cssContent)
	})

	mux.HandleFunc("/app.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(iconBytes)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	server = &http.Server{
		Handler: mux,
	}

	// 분리된 browser.go의 함수를 호출하여 메인 루프를 간결하게 유지
	go func() {
		time.Sleep(50 * time.Millisecond)
		launchAppWindow(url, 520, 450)
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Println(err)
	}
}

// 💡 [개선] html/template 오버헤드를 제거하고 완전한 static 스트링 구조로 직통 서빙
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, htmlContent)
}
