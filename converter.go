// converter.go
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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

// 대용량 유입에도 메모리를 먹지 않는 완전한 스트리밍 변환기
func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "잘못된 접근입니다.", http.StatusMethodNotAllowed)
		return
	}

	// 최대 100MB 업로드 허용
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "업로드 파일이 너무 크거나 손상되었습니다.", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("csvFile")
	if err != nil {
		http.Error(w, "파일을 읽는 중 오류가 발생했습니다.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 1. 헤더 한 줄 먼저 읽기
	headers, err := reader.Read()
	if err != nil {
		http.Error(w, "헤더를 읽을 수 없습니다.", http.StatusBadRequest)
		return
	}

	for i := range headers {
		headers[i] = removeBOM(headers[i])
	}

	if err := validateHeaders(headers); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. 다운로드 헤더 설정 (w.Write 전에 무조건 먼저 선언해야 함)
	downloadName := header.Filename
	if strings.HasSuffix(strings.ToLower(downloadName), ".csv") {
		downloadName = downloadName[:len(downloadName)-4] + ".json"
	} else {
		downloadName += ".json"
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadName))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 3. 브라우저로 직접 스트리밍 출력 시작 (Buffer 완전히 제거)
	io.WriteString(w, "[\n")

	firstRecord := true
	recordCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// 이미 헤더가 전송된 이후이므로 http.Error 대신 에러 로그 처리나 스트림 중단만 가능
			return
		}

		recordCount++

		if !firstRecord {
			io.WriteString(w, ",\n")
		}
		firstRecord = false

		io.WriteString(w, "    {\n")

		for i := 0; i < len(headers); i++ {
			value := ""
			if i < len(record) {
				value = record[i]
			}

			// 안전하고 정확한 표준 json 마샬러 활용 (escape 함수 필요 없음)
			keyJSON, _ := json.Marshal(headers[i])
			valueJSON, _ := json.Marshal(value)

			fmt.Fprintf(w, "        %s: %s", string(keyJSON), string(valueJSON))

			if i < len(headers)-1 {
				io.WriteString(w, ",\n")
			} else {
				io.WriteString(w, "\n")
			}
		}
		io.WriteString(w, "    }")
	}

	// 데이터가 아예 없었던 경우 처리
	if recordCount == 0 {
		// 주의: 이미 200 OK 상태로 데이터가 일부 나갔을 수 있으므로 빈 배열로 마감
		io.WriteString(w, "]\n")
		return
	}

	io.WriteString(w, "\n]")
}
