// converter.go
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

// 💡 각 세포(Cell)의 문자를 분석하여 알맞은 JSON 데이터 타입 표현식으로 변환
func inferValueJSON(s string) string {
	sClean := strings.TrimSpace(s)

	// 1. 빈 값은 그냥 빈 문자열로 처리
	if sClean == "" {
		return `""`
	}

	// 2. 불리언(Boolean) 판별
	if sClean == "true" || sClean == "false" {
		return sClean
	}

	// 1️⃣ "010", "00234" 등 앞자리 0 전면 방어
	if len(sClean) > 1 && sClean[0] == '0' && sClean[1] != '.' {
		b, _ := json.Marshal(s) // 숫자로 바꾸지 않고 곧바로 "010" 문자열 처리
		return string(b)
	}

	// 2️⃣ 변환 시도하다가 에러 나면?
	if _, err := strconv.ParseInt(sClean, 10, 64); err == nil {
		return sClean
	}
	if _, err := strconv.ParseFloat(sClean, 64); err == nil {
		if !strings.Contains(sClean, "NaN") && !strings.Contains(sClean, "Inf") {
			return sClean
		}
	}

	// 3️⃣ [최종 방어선] 에러가 나서 여기까지 흘러오면 무조건 문자열 처리!
	b, _ := json.Marshal(s)
	return string(b)
}

// 대용량 유입에도 메모리를 먹지 않는 완전한 스트리밍 변환기
func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "잘못된 접근입니다.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "업로드 파일이 너무 크거나 손상되었습니다.", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csvFile") // 👈 사용하지 않는 header 식별자 생략 (_)
	if err != nil {
		http.Error(w, "파일을 읽는 중 오류가 발생했습니다.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

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

	// 👈 [개선] 파일명은 프론트엔드가 핸들링하므로, 백엔드는 표준 다운로드 헤더 스펙만 깔끔하게 유지합니다.
	w.Header().Set("Content-Disposition", "attachment; filename=\"download.json\"")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	io.WriteString(w, "[\n")

	firstRecord := true
	recordCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
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

			// 헤더(Key)는 무조건 문자열이므로 표준 마샬러 사용
			keyJSON, _ := json.Marshal(headers[i])

			// 데이터 세포별로 타입을 자동 추정하여 주입
			valueJSON := inferValueJSON(value)

			fmt.Fprintf(w, "        %s: %s", string(keyJSON), valueJSON)

			if i < len(headers)-1 {
				io.WriteString(w, ",\n")
			} else {
				io.WriteString(w, "\n")
			}
		}
		io.WriteString(w, "    }")
	}

	if recordCount == 0 {
		io.WriteString(w, "]\n")
		return
	}

	io.WriteString(w, "\n]")
}
