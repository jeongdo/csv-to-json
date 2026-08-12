# CSV to JSON

Go + Wails로 만든 가볍고 로컬 우선인 Windows 데스크톱 CSV → JSON 변환기입니다.

CSV/TSV/구분자 텍스트 파일을 외부로 업로드하지 않고 PC 안에서 JSON으로 변환합니다. 화면단은 파일 내용이 아니라 **파일 경로만 Go에 전달**하고, Go가 원본 파일을 직접 열어 레코드 단위로 처리합니다. 따라서 대용량 CSV도 브라우저 Blob이나 전체 메모리 적재 없이 처리할 수 있습니다.

## 주요 기능

- Wails 네이티브 데스크톱 창 — localhost 서버 / Chrome·Edge App Mode 제거
- Drag & Drop + 네이티브 파일 열기/저장 대화상자
- `encoding/csv` 기반 직접 디스크 스트리밍
- 제한된 미리보기: 헤더 + 최초 8행만 읽기
- 구분자 자동 감지: 쉼표, 탭, 파이프, 세미콜론
- 1개 컬럼 CSV 정상 지원
- UTF-8 BOM 헤더 자동 제거
- 빈 헤더 / 중복 헤더 검증
- CSV 컬럼 순서 그대로 JSON key 순서 유지
- JSON 데이터 타입 자동 추론 On/Off
- 빈 셀 → `null` 옵션
- `01001` 같은 선행 0 식별자 문자열 보호
- `9223372036854775808` 같은 큰 숫자도 `float64` 정밀도 손실 없이 그대로 출력
- 문자열 앞뒤 공백 보존
- 대용량 파일 진행률 표시
- 안전한 결과 저장: 임시 파일에 끝까지 성공한 뒤에만 최종 JSON 교체
- 한국어 / 영어 UI
- 라이트 / 다크 테마
- 완전 로컬 처리

## 데이터 안전 구조

```text
CSV / TSV 파일
    │
    ├─ 검사: 헤더 + 최대 8행 미리보기
    │
    └─ 변환
         │
         ▼
    Go가 원본 파일 직접 Open
         │
         ▼
    csv.Reader 레코드 단위 처리
         │
         ▼
    임시 JSON 파일 작성
         │
         ├─ 실패 → 임시 파일 삭제 / 기존 결과 파일 유지
         │
         └─ 성공 → 최종 JSON 파일로 교체
```

변환 과정에서 프런트엔드는 `File.arrayBuffer()`, `File.text()`, `response.blob()` 또는 HTTP 업로드를 사용하지 않습니다.

## 타입 추론 예시

**데이터 타입 자동 추론**을 켠 경우:

```csv
name,age,zipcode,active,big
Kim,20,01001,true,9223372036854775808
```

결과:

```json
[
  {
    "name": "Kim",
    "age": 20,
    "zipcode": "01001",
    "active": true,
    "big": 9223372036854775808
  }
]
```

`+1`, `.5`, `1.`, `-01`처럼 JSON 숫자 문법에 맞지 않는 값은 억지로 숫자로 만들지 않고 문자열로 유지하므로 깨진 JSON을 만들지 않습니다.

## 프로젝트 구조

```text
.
├─ main.go                  # Wails 앱 진입점
├─ app.go                   # 네이티브 대화상자 / UI 바인딩
├─ converter.go             # 순수 CSV → JSON 변환 코어
├─ file_io.go               # 직접 파일 스트리밍 / 안전한 결과 교체
├─ converter_test.go
├─ file_io_test.go
├─ frontend/
│  └─ dist/
│     ├─ index.html
│     ├─ app.js
│     └─ style.css
├─ build/
│  └─ windows/
│     └─ icon.ico
├─ .github/workflows/
│  └─ windows-build.yml
└─ wails.json
```

## 빌드

개발 환경:

- Go 1.26+
- Wails v2.13.0
- Windows exe 빌드 시 Windows 환경

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
go test .
wails build -clean
```

Windows 실행 파일:

```text
build/bin/CsvToJson.exe
```

`main`에 push될 때마다 GitHub Actions에서 Go 테스트와 Windows exe 빌드를 자동 검증합니다.

## 현재 범위

의도적으로 아직 넣지 않은 기능:

- XLSX 입력
- Nested JSON 생성
- 스키마 매핑/편집
- 임의 출력 템플릿

이 프로젝트는 기능을 무작정 늘리기보다 **CSV → JSON 하나를 빠르고 안전하고 믿을 수 있게 처리하는 작은 데스크톱 도구**를 목표로 합니다.

## 라이선스

Apache License 2.0. 자세한 내용은 [LICENSE](LICENSE)를 참고하세요.
