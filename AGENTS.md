# Agent Rules

## GitHub Actions 사용 정책

이 프로젝트는 GitHub Actions를 기본 검증·빌드 환경으로 사용하지 않는다.

GitHub Actions는 현재 Agent 실행환경에서 수행할 수 없는 작업을 보완하거나, 사용자가 최종 산출물을 ChatGPT 대화에서 바로 다운로드할 수 있도록 artifact를 생성해야 하는 경우에만 제한적으로 사용한다.

### 1. 기본 원칙: Agent가 직접 실행하고 검증한다

현재 Agent 실행환경에서 가능한 작업은 GitHub Actions로 넘기지 않는다.

예:

- Python: pytest, ruff, compile, smoke
- Go: go test, go vet, build
- Node/React: npm test, npm run build, lint
- Rust: cargo test, cargo clippy, cargo build
- Java: Gradle/Maven test/build
- ZIP/bundle 생성
- 데이터 검증 및 정적 검사

코드 수정 후 가능한 검증은 Agent가 직접 수행하고, 검증 완료 후 commit/push 한다. 단순 lint/test/build를 위해 GitHub Actions를 사용하지 않는다.

### 2. Windows / Android도 Actions가 기본이 아니다

Windows 또는 Android 프로젝트라는 이유만으로 GitHub Actions를 사용하지 않는다. Agent가 현재 환경에서 필요한 검증을 수행할 수 있다면 우선 직접 검증한다.

Windows EXE, Android APK/AAB 등의 최종 산출물도 사용자가 자신의 개발환경에서 직접 빌드할 수 있다면 Actions 실행은 필수가 아니다.

### 3. 다운로드 가능한 최종 산출물이 필요할 때 Actions 사용 가능

사용자가 다음과 같이 명시적으로 요청한 경우에는 GitHub Actions를 최종 artifact 생성 수단으로 사용할 수 있다.

예:

- "exe 만들어서 여기서 받을 수 있게 해줘"
- "apk 만들어서 채팅에서 다운로드하게 해줘"
- "Windows 실행파일 줘"
- "최종 빌드 결과물을 artifact로 만들어줘"

이 경우 Agent가 먼저 가능한 검증을 수행한 뒤 GitHub Actions의 특수 빌드 환경에서 EXE/APK/AAB 등의 artifact를 생성하고 결과를 확인한다.

즉 Windows/Android Actions의 주 목적은 일상적인 CI가 아니라 사용자가 바로 받을 수 있는 최종 실행파일을 만드는 것이다.

### 4. 특수 실행환경이 정말 필요한 경우

다음과 같이 현재 Agent 실행환경에서 직접 수행하기 어려운 경우에도 GitHub Actions 사용을 검토할 수 있다.

- Windows 네이티브 빌드
- 특정 OS 전용 빌드
- 여러 OS의 실제 네이티브 결과 비교
- GitHub Release용 최종 artifact 생성

가능한 경우에도 필요할 때만 실행한다.

### 5. GPU 작업은 GitHub Actions 대상이 아니다

모델 학습이나 GPU 검증은 기본적으로 GitHub Actions로 처리하지 않는다. GPU가 필요한 작업은 프로젝트에서 지정한 Google Colab, 전용 GPU 서버, 사용자 GPU 환경 등을 사용한다. CPU에서 가능한 preflight/gate/test는 Agent가 직접 수행한다.

### 6. 자동 실행보다 필요 시 실행을 우선한다

Actions가 필요한 프로젝트라도 기본적으로 `push`마다 자동 실행하는 것보다 `workflow_dispatch` 등 필요할 때 실행하는 방식을 우선한다. 특히 EXE/APK/AAB 같은 최종 artifact는 모든 commit마다 생성할 필요가 없다.

### 7. 기존 workflow가 있다는 이유만으로 사용하지 않는다

`.github/workflows/*`가 존재한다고 해서 해당 Actions가 현재도 필요한 것으로 간주하지 않는다. Agent는 workflow가 수행하는 작업을 먼저 확인하고, 동일한 작업을 현재 실행환경에서 수행할 수 있다면 직접 검증을 우선한다.

불필요한 Actions workflow를 새로 만들거나 확대하지 않는다. 기존 workflow의 정리·삭제는 별도 작업 범위에서 판단하며 이 정책을 이유로 무조건 삭제하지 않는다.

### 8. Actions 때문에 저장소 공개 범위를 변경하지 않는다

Actions 비용이나 편의를 이유로 Private repository를 Public으로 변경하지 않는다. 저장소 공개 여부는 프로젝트와 소스의 공개 적합성을 기준으로 판단한다.

### 9. Actions 인프라 실패와 코드 실패를 구분한다

결제, spending limit, runner 할당, Actions 서비스 문제 등으로 job이 실제 코드 실행 전에 실패한 경우 프로젝트 코드 실패로 간주하지 않는다. 가능한 로컬/Agent 검증으로 작업을 계속한다.

### 핵심 판단 순서

1. 현재 Agent가 직접 실행할 수 있는가? 가능하면 직접 실행하고 검증한다.
2. 사용자가 채팅에서 다운로드할 최종 산출물을 명시적으로 요청했는가? 그렇다면 필요한 경우 Actions artifact를 사용한다.
3. 현재 Agent에 없는 특수 OS/환경이 실제로 필요한가? 그렇다면 제한적으로 Actions를 사용한다.
4. 위 조건에 해당하지 않으면 Actions를 실행하거나 새로 추가하지 않는다.
