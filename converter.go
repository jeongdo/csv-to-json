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

	// [FIX 3] 앞자리 0 방어 — json.Marshal 대상을 s 에서 sClean 으로 통일
	// "010", "00234" 등 앞자리 0 전면 방어
	if len(sClean) > 1 && sClean[0] == '0' && sClean[1] != '.' {
		b, _ := json.Marshal(sClean)
		return string(b)
	}

	if _, err := strconv.ParseInt(sClean, 10, 64); err == nil {
		return sClean
	}

	if _, err := strconv.ParseFloat(sClean, 64); err == nil {
		if !strings.Contains(sClean, "NaN") && !strings.Contains(sClean, "Inf") {
			return sClean
		}
	}

	// [FIX 3] 최종 방어선도 sClean 으로 통일 (원본 s 의 앞뒤 공백이 JSON에 포함되던 버그 수정)
	b, _ := json.Marshal(sClean)
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

	file, _, err := r.FormFile("csvFile")
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

	delim, mixed := detectDelimiterAdvanced(file)
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

		io.WriteString(w, "  {\n")

		for i := 0; i < len(headers); i++ {
			value := ""
			if i < len(record) {
				value = record[i]
			}

			keyJSON, _ := json.Marshal(headers[i])
			valueJSON := inferValueJSON(value)

			fmt.Fprintf(w, "    %s: %s", string(keyJSON), valueJSON)

			if i < len(headers)-1 {
				io.WriteString(w, ",\n")
			} else {
				io.WriteString(w, "\n")
			}
		}

		io.WriteString(w, "  }")
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

	// 끊긴 마지막 줄 제외
	if len(lines) > 1 && n == len(buf) {
		lines = lines[:len(lines)-1]
	}

	// 헤더 줄이 비어있으면 오류
	if len(lines) == 0 {
		return ',', true
	}

	candidates := []rune{',', '|', ';', '\t'}

	// [FIX 1] sample 전체가 아닌 lines[0](헤더 줄)에서만 후보 추림
	// 데이터 셀 안에 |, ;, \t 가 포함되어도 오탐하지 않음
	headerLine := lines[0]
	var presentCandidates []rune
	for _, d := range candidates {
		if strings.ContainsRune(headerLine, d) {
			presentCandidates = append(presentCandidates, d)
		}
	}

	// 어떤 구분자도 없다면 잘못된 파일
	if len(presentCandidates) == 0 {
		return ',', true
	}

	// [FIX 2] lines 를 줄별로 독립 파싱하지 않고 전체를 하나의 csv.Reader 로 파싱
	// → 따옴표 안 줄바꿈/쉼표가 있는 RFC 4180 파일도 정확히 칼럼 수 검증
	var validDelims []rune
	for _, d := range presentCandidates {
		if isValidDelimiter(lines, d) {
			validDelims = append(validDelims, d)
		}
	}

	if len(validDelims) == 0 {
		return ',', true
	}

	if len(validDelims) > 1 {
		return ',', true
	}

	return validDelims[0], false
}

// [FIX 2] lines 를 다시 합쳐서 단일 csv.Reader 로 전체 파싱
// 줄별 독립 파싱 시 따옴표 내 줄바꿈을 칼럼 수 불일치로 오판하던 버그 수정
func isValidDelimiter(lines []string, delim rune) bool {
	joined := strings.Join(lines, "\n")
	r := csv.NewReader(strings.NewReader(joined))
	r.Comma = delim
	r.FieldsPerRecord = 0 // 첫 레코드 칼럼 수를 기준으로 자동 설정
	r.LazyQuotes = true   // 따옴표 파싱 관대하게

	var baseLen int
	found := false

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false
		}
		if !found {
			baseLen = len(rec)
			found = true
			continue
		}
		if len(rec) != baseLen {
			return false
		}
	}

	return found
}
