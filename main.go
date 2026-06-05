package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

//go:embed index.html
var htmlContent string

//go:embed csv_to_json.ico
var iconBytes []byte

var server *http.Server

const (
	smCxScreen = 0
	smCyScreen = 1
)

func getScreenResolution() (int, int) {
	if runtime.GOOS != "windows" {
		return 1920, 1080
	}

	mod := syscall.NewLazyDLL("user32.dll")
	proc := mod.NewProc("GetSystemMetrics")

	width, _, _ := proc.Call(uintptr(smCxScreen))
	height, _, _ := proc.Call(uintptr(smCyScreen))

	return int(width), int(height)
}

func killExistingProcesses() {
	currentPID := os.Getpid()

	if runtime.GOOS == "windows" {

		cmd := exec.Command(
			"taskkill",
			"/F",
			"/IM",
			"CsvToJson.exe",
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

	w.WriteHeader(http.StatusOK)

	go func() {

		time.Sleep(500 * time.Millisecond)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)

		defer cancel()

		server.Shutdown(ctx)

		os.Exit(0)
	}()
}

func removeBOM(s string) string {
	return strings.TrimPrefix(s, "\uFEFF")
}

func validateHeaders(headers []string) error {

	exists := make(map[string]bool)

	for _, h := range headers {

		h = strings.TrimSpace(h)

		if h == "" {
			return fmt.Errorf("빈 헤더가 존재합니다")
		}

		if exists[h] {
			return fmt.Errorf("중복 헤더 발견: %s", h)
		}

		exists[h] = true
	}

	return nil
}

func main() {

	go killExistingProcesses()

	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/convert", convertHandler)
	mux.HandleFunc("/shutdown", shutdownHandler)

	mux.HandleFunc("/app.ico", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(iconBytes)
	})

	listener, err := net.Listen(
		"tcp",
		"127.0.0.1:0",
	)

	if err != nil {
		return
	}

	port := listener.Addr().(*net.TCPAddr).Port

	url := fmt.Sprintf(
		"http://127.0.0.1:%d",
		port,
	)

	server = &http.Server{
		Handler: mux,
	}

	go func() {

		time.Sleep(50 * time.Millisecond)

		scrWidth, scrHeight := getScreenResolution()

		appWidth := 520
		appHeight := 450

		posX := (scrWidth - appWidth) / 2
		posY := (scrHeight - appHeight) / 2

		var cmd *exec.Cmd

		switch runtime.GOOS {

		case "windows":

			sizeArg := fmt.Sprintf(
				"--window-size=%d,%d",
				appWidth,
				appHeight,
			)

			posArg := fmt.Sprintf(
				"--window-position=%d,%d",
				posX,
				posY,
			)

			cmd = exec.Command(
				"cmd",
				"/c",
				"start",
				"chrome",
				"--app="+url,
				sizeArg,
				posArg,
			)

			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}

		case "darwin":

			cmd = exec.Command(
				"open",
				"-a",
				"Google Chrome",
				"--args",
				"--app="+url,
			)

		default:

			cmd = exec.Command(
				"xdg-open",
				url,
			)
		}

		cmd.Run()
	}()

	server.Serve(listener)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index").Parse(htmlContent)

	if err != nil {
		http.Error(
			w,
			"템플릿 로딩 실패",
			http.StatusInternalServerError,
		)
		return
	}

	tmpl.Execute(w, nil)
}

func convertHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"잘못된 접근입니다.",
			http.StatusMethodNotAllowed,
		)
		return
	}

	r.ParseMultipartForm(100 << 20)

	file, header, err := r.FormFile("csvFile")

	if err != nil {
		http.Error(
			w,
			"파일을 읽는 중 오류가 발생했습니다.",
			http.StatusBadRequest,
		)
		return
	}

	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()

	if err != nil {
		http.Error(
			w,
			"헤더를 읽을 수 없습니다.",
			http.StatusBadRequest,
		)
		return
	}

	for i := range headers {
		headers[i] = removeBOM(headers[i])
	}

	if err := validateHeaders(headers); err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	var jsonBuf bytes.Buffer

	jsonBuf.WriteString("[\n")

	firstRecord := true
	recordCount := 0

	for {

		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {

			http.Error(
				w,
				"CSV 파싱 실패: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		recordCount++

		if !firstRecord {
			jsonBuf.WriteString(",\n")
		}

		firstRecord = false

		jsonBuf.WriteString("    {\n")

		for i := 0; i < len(headers); i++ {

			value := ""

			if i < len(record) {
				value = record[i]
			}

			keyJSON, _ := json.Marshal(headers[i])
			valueJSON, _ := json.Marshal(value)

			jsonBuf.WriteString(
				fmt.Sprintf(
					"        %s: %s",
					string(keyJSON),
					string(valueJSON),
				),
			)

			if i < len(headers)-1 {
				jsonBuf.WriteString(",\n")
			} else {
				jsonBuf.WriteString("\n")
			}
		}

		jsonBuf.WriteString("    }")
	}

	if recordCount == 0 {

		http.Error(
			w,
			"데이터가 부족합니다. (최소 헤더 1행, 데이터 1행 필요)",
			http.StatusBadRequest,
		)

		return
	}

	jsonBuf.WriteString("\n]")

	downloadName := header.Filename

	if strings.HasSuffix(
		strings.ToLower(downloadName),
		".csv",
	) {

		downloadName =
			downloadName[:len(downloadName)-4] + ".json"

	} else {

		downloadName += ".json"
	}

	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=%s",
			downloadName,
		),
	)

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.Write(jsonBuf.Bytes())
}
