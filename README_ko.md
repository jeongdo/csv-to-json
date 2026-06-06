# csv-to-json

Go로 개발한 경량 데스크톱 CSV → JSON 변환기

csv-to-json은 CSV 파일을 JSON 형식으로 변환하는 포터블(Portable) 데스크톱 유틸리티입니다.

간단한 그래픽 UI를 제공하며, 별도 설치 없이 실행 가능하고 외부 의존성이 없습니다.

---

## 주요 기능

* 단일 실행 파일(Exe) 배포
* 자동 구분자 감지 (CSV, TSV, 파이프, 세미콜론)
* Go Embed를 이용한 HTML, CSS, JavaScript, 아이콘 내장
* 로컬 전용 처리 (외부 네트워크 통신 없음)
* JSON 자동 다운로드
* CSV 헤더 검증
* 중복 헤더 검출
* 빈 헤더 검출
* UTF-8 BOM 자동 제거
* CSV 컬럼 순서 유지
* JSON 데이터 타입 자동 추론
* 선행 0(Leading Zero) 보호
* 스트리밍 기반 CSV 변환 (메모리 사용 최소화)
* 한국어 / 영어 UI 지원
* Chrome / Edge App Mode 기반 데스크톱 UI
* Toast 알림 기반 오류 처리
* CSV 파일 Drag & Drop 지원
* 프로그램 종료 시 자동 프로세스 정리

---

## 동작 방식

```text
CSV 파일
    ↓
csv-to-json
    ↓
CSV 검증
    ↓
스트리밍 변환
    ↓
JSON 다운로드
```

프로그램은 임시 로컬 HTTP 서버를 실행한 후 브라우저를 App Mode로 실행합니다.

CSV 데이터를 JSON으로 변환한 뒤 자동으로 다운로드를 시작합니다.

모든 처리는 사용자 PC 내부에서만 수행됩니다.

---

## 지원 데이터 타입

csv-to-json은 입력된 데이터를 분석하여 최적의 JSON 데이터 타입을 자동으로 적용합니다.

지원 타입:

* 문자열(String)
* 정수(Integer)
* 실수(Float)
* 불리언(Boolean)

### 입력 CSV

```csv
name,age,salary,active
Kim,20,3500.50,true
Lee,30,4200,false
```

### 출력 JSON

```json
[
  {
    "name": "Kim",
    "age": 20,
    "salary": 3500.50,
    "active": true
  },
  {
    "name": "Lee",
    "age": 30,
    "salary": 4200,
    "active": false
  }
]
```

---

## 선행 0 보호

### 입력

```csv
zipcode
01001
```

### 출력

```json
{
  "zipcode": "01001"
}
```

우편번호, 사번, 상품코드 등 선행 0이 의미를 가지는 값은 문자열로 유지됩니다.

이를 통해 데이터 손상을 방지합니다.

---

## CSV 검증

다음 항목을 자동 검증합니다.

* 빈 헤더
* 중복 헤더
* 헤더 누락
* CSV 형식 오류

### 잘못된 CSV 예시

```csv
name,name,age
Kim,20,Seoul
```

### 결과

```text
중복 헤더 발견: name
```

---

## 아키텍처

```text
CsvToJson.exe
        │
        ▼
내장 HTTP 서버
(127.0.0.1 임의 포트)
        │
        ▼
Chrome / Edge App Mode
        │
        ▼
HTML + CSS + JavaScript UI
        │
        ▼
스트리밍 CSV 파서
        │
        ▼
JSON 다운로드
```

---

## 브라우저 지원

Windows 실행 우선순위:

1. Google Chrome
2. Microsoft Edge
3. 기본 브라우저

Chrome 또는 Edge가 설치되어 있지 않은 경우 자동으로 기본 브라우저를 사용합니다.

---

## 프로젝트 구조

```text
csv-to-json
│
├── main.go
├── browser.go
├── converter.go
├── index.html
├── style.css
├── app.js
├── app.ico
└── README.md
```

---

## 빌드

### 1. 사전 준비 (필수 도구 설치)

아이콘 및 Windows 리소스 처리를 위해 `goversioninfo`를 설치합니다.

```bash
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

### 2. 리소스 파일 확인 (선택 사항)

실행 파일에 아이콘과 버전 정보를 포함하려면 프로젝트 루트에 다음 파일들을 준비합니다.

* app.ico : 실행 파일 아이콘
* versioninfo.json : 버전 정보 설정 파일
* app.manifest : Windows 실행 권한 매니페스트

### 3. 리소스 생성 (.syso 파일 생성)

```bash
goversioninfo -platform-specific=true
```

### 4. 최종 빌드

```bash
go build -ldflags="-H windowsgui" -o CsvToJson.exe
```

---

## 요구사항

* Windows 10 이상
* Go 1.26.4 (개발 시에만 필요)

최종 사용자는 별도 설치가 필요하지 않습니다.

---

## 제한 사항

현재 버전은 단순성과 경량성에 초점을 맞추고 있습니다.

지원하지 않는 기능:

* XLSX 입력
* 중첩(Nested) JSON 생성
* 스키마 매핑
* 사용자 정의 출력 포맷

---

## 라이선스

MIT License

자유롭게 사용, 수정 및 배포할 수 있습니다.
