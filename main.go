package main

import (
	"bytes"
	"context"
	_ "embed" // Go 공식 내장 임베딩 시스템
	"encoding/csv"
	"fmt"
	"html/template"
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

// 프로젝트 폴더 내에 csv_to_json.ico 파일이 존재하면 바이너리 메모리 안에 함께 임베딩합니다.
// 만약 아이콘 파일이 없다면 해당 줄과 아래 iconBytes 변수 서빙 핸들러를 주석 처리하세요.
//
//go:embed csv_to_json.ico
var iconBytes []byte

var server *http.Server

const (
	smCxScreen = 0
	smCyScreen = 1
)

// 윈도우 모니터 해상도를 획득하기 위한 Win32 API
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

// 중복 실행 감지 시 기존에 메모리에 떠 있는 구버전 CsvToJson.exe 프로세스를 강제 청소하는 기능
func killExistingProcesses() {
	currentPID := os.Getpid()
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", "CsvToJson.exe", "/FI", fmt.Sprintf("PID ne %d", currentPID))
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Run()
		time.Sleep(200 * time.Millisecond)
	}
}

// 창이 닫힐 때 Go 프로세스를 메모리상에서 완벽히 영제하는 종료 헨들러
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
	// 킬러 로직을 비동기로 할당하여 바탕화면 더블 클릭 체감 속도를 극대화
	go killExistingProcesses()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/convert", convertHandler)
	mux.HandleFunc("/shutdown", shutdownHandler)

	// 내장 아이콘 데이터 개설 핸들러
	mux.HandleFunc("/app.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(iconBytes)
	})

	// 남는 빈 포트를 OS가 실시간으로 할당하게 조치 (중복 실행 포트 충돌 전면 차단)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	server = &http.Server{Handler: mux}

	// 주소창 없는 크롬 단독 화면 앱 실행 루프
	go func() {
		time.Sleep(50 * time.Millisecond)

		scrWidth, scrHeight := getScreenResolution()

		// 초기 안착 팝업 사양 설정 및 모니터 한가운데 정확한 기하학적 정중앙 계산
		appWidth := 520
		appHeight := 450
		posX := (scrWidth - appWidth) / 2
		posY := (scrHeight - appHeight) / 2

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			sizeArg := fmt.Sprintf("--window-size=%d,%d", appWidth, appHeight)
			posArg := fmt.Sprintf("--window-position=%d,%d", posX, posY)
			cmd = exec.Command("cmd", "/c", "start", "chrome", "--app="+url, sizeArg, posArg)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		case "darwin":
			cmd = exec.Command("open", "-a", "Google Chrome", "--args", "--app="+url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		cmd.Run()
	}()

	server.Serve(listener)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("index").Parse(htmlContent)
	tmpl.Execute(w, nil)
}

// 가볍고 정교한 바이트 스트리밍 파싱 코어 변환부
func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "잘못된 접근입니다.", http.StatusMethodNotAllowed)
		return
	}

	// 대용량 유입 대비 서버 인프라 보호용 버퍼 규격 선언 (최대 100MB 허용)
	r.ParseMultipartForm(100 << 20)

	file, header, err := r.FormFile("csvFile")
	if err != nil {
		http.Error(w, "파일을 읽는 중 오류가 발생했습니다.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "CSV 파싱 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(records) < 2 {
		http.Error(w, "데이터가 부족합니다. (최소 헤더 1행, 데이터 1행 필요)", http.StatusBadRequest)
		return
	}

	// 첫 번째 줄 전체를 고정 칼럼 헤더 배열로 박제
	headers := records[0]

	// 중가 무거운 데이터 매핑 구조 없이 연속된 메모리에 바이트 스트림으로 JSON 문자열 강제 결합
	var jsonBuf bytes.Buffer
	jsonBuf.WriteString("[\n")

	for rowIndex, row := range records[1:] {
		jsonBuf.WriteString("    {\n")
		for i, cell := range row {
			if i >= len(headers) {
				break
			}
			escapedKey := escapeJSONString(headers[i])
			escapedValue := escapeJSONString(cell)
			jsonBuf.WriteString(fmt.Sprintf("        \"%s\": \"%s\"", escapedKey, escapedValue))

			if i < len(row)-1 && i < len(headers)-1 {
				jsonBuf.WriteString(",\n")
			} else {
				jsonBuf.WriteString("\n")
			}
		}

		if rowIndex < len(records)-2 {
			jsonBuf.WriteString("    },\n")
		} else {
			jsonBuf.WriteString("    }\n")
		}
	}
	jsonBuf.WriteString("]")

	downloadName := header.Filename
	if strings.HasSuffix(strings.ToLower(downloadName), ".csv") {
		downloadName = downloadName[:len(downloadName)-4] + ".json"
	} else {
		downloadName = downloadName + ".json"
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadName))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonBuf.Bytes())
}

// 텍스트 깨짐 및 JSON 이스케이프 파괴 현상 방어 장치
func escapeJSONString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
