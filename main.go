// main.go

package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
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

// 앱 기동 시 랜덤 생성 — 프로세스 재시작 전까지 동일한 값 유지
var shutdownToken string

func generateShutdownToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback: 시간 기반 난수 (rand.Read 실패는 사실상 없음)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func killExistingProcesses() {
	currentPID := os.Getpid()

	if runtime.GOOS == "windows" {
		exePath, err := os.Executable()
		if err != nil {
			return
		}
		exeName := filepath.Base(exePath)
		cmd := exec.Command(
			"taskkill",
			"/F",
			"/IM",
			exeName,
			"/FI",
			fmt.Sprintf("PID ne %d", currentPID),
		)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		cmd.Run()
		time.Sleep(200 * time.Millisecond)
	}
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	// [SECURITY] 토큰 검증 — 없거나 틀리면 403
	if r.URL.Query().Get("token") != shutdownToken {
		w.WriteHeader(http.StatusForbidden)
		return
	}

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
	shutdownToken = generateShutdownToken()

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

	// [SECURITY] 토큰을 URL에 포함시켜 브라우저로 전달
	url := fmt.Sprintf("http://127.0.0.1:%d?token=%s", port, shutdownToken)

	server = &http.Server{
		Handler: mux,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		launchAppWindow(url, 520, 450)
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Println(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, htmlContent)
}
