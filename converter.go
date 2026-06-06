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

type ErrorResponse struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	detail string,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		ErrorResponse{
			Code:   code,
			Detail: detail,
		},
	)
}

func removeBOM(s string) string {
	return strings.TrimPrefix(s, "\uFEFF")
}

func validateHeaders(headers []string) error {

	exists := make(map[string]bool)

	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			return fmt.Errorf("EMPTY_HEADER")
		}

		if exists[h] {
			return fmt.Errorf(
				"DUPLICATE_HEADER:%s",
				h,
			)
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
		writeError(
			w,
			http.StatusMethodNotAllowed,
			"INVALID_METHOD",
			"",
		)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"UPLOAD_FAILED",
			"",
		)
		return
	}

	file, _, err := r.FormFile("csvFile") // 👈 사용하지 않는 header 식별자 생략 (_)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"FILE_READ_FAILED",
			"",
		)
		return
	}
	defer file.Close()

	buf := make([]byte, 512*1024)
	n, _ := file.Read(buf)

	delim, mixed := detectDelimiterAdvanced(strings.NewReader(string(buf[:n])))

	if mixed {
		writeError(
			w,
			http.StatusBadRequest,
			"MIXED_DELIMITER_DETECTED",
			"",
		)
		return
	}

	// 안전하게 rewind (Seek 가능한 경우만)
	if seeker, ok := file.(io.Seeker); ok {
		_, _ = seeker.Seek(0, 0)
	}

	reader := csv.NewReader(file)
	reader.Comma = delim

	headers, err := reader.Read()
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"HEADER_READ_FAILED",
			"",
		)
		return
	}

	for i := range headers {
		headers[i] = removeBOM(headers[i])
	}

	if err := validateHeaders(headers); err != nil {
		msg := err.Error()

		if msg == "EMPTY_HEADER" {
			writeError(
				w,
				http.StatusBadRequest,
				"EMPTY_HEADER",
				"",
			)
			return
		}

		if strings.HasPrefix(
			msg,
			"DUPLICATE_HEADER:",
		) {
			writeError(
				w,
				http.StatusBadRequest,
				"DUPLICATE_HEADER",
				strings.TrimPrefix(
					msg,
					"DUPLICATE_HEADER:",
				),
			)
			return
		}
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

func detectDelimiterAdvanced(r io.Reader) (rune, bool) {

	buf := make([]byte, 512*1024)
	n, _ := r.Read(buf)
	sample := string(buf[:n])

	lines := strings.Split(sample, "\n")

	candidates := []rune{',', '|', ';', '\t'}

	// 1. 텍스트 안에 실제로 존재하는 구분자만 후보로 추려냅니다.
	var presentCandidates []rune
	for _, d := range candidates {
		if strings.ContainsRune(sample, d) {
			presentCandidates = append(presentCandidates, d)
		}
	}

	// 2. 어떤 구분자도 없다면 잘못된 파일로 간주하고 에러 반환
	if len(presentCandidates) == 0 {
		return ',', true
	}

	// 3. 존재하는 구분자들만 대상으로 칼럼 일관성 검사
	var validDelims []rune
	for _, d := range presentCandidates {
		// 💡 targetLines 대신 원본 lines를 그대로 전달합니다.
		if isValidDelimiter(lines, d) {
			validDelims = append(validDelims, d)
		}
	}

	// 존재하는 구분자들 중 구조가 맞는 게 없으면 파일 손상
	if len(validDelims) == 0 {
		return ',', true
	}

	// 통과한 구분자가 여러 개면 진짜 혼합 에러
	if len(validDelims) > 1 {
		return ',', true
	}

	return validDelims[0], false
}

func isValidDelimiter(lines []string, delim rune) bool {

	var baseLen int
	found := false

	for _, line := range lines {

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		r := csv.NewReader(strings.NewReader(line))
		r.Comma = delim

		rec, err := r.Read()
		if err != nil {
			return false
		}

		if !found {
			baseLen = len(rec)
			found = true
			continue
		}

		// 구조 깨지면 즉시 탈락
		if len(rec) != baseLen {
			return false
		}
	}

	return found
}
