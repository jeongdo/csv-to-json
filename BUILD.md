# Build / Test Policy

이 문서는 이 저장소의 코드 변경 후 빌드·테스트 검증 정책이다.

`AGENTS.md`가 이 문서를 참조하며, 과거 문서나 지침에 남아 있는 GitHub Actions 제한/금지 규칙과 충돌하면 **이 문서를 우선한다.**

## 1. 작업선 확인

- `main`을 고정으로 가정하지 않는다. 저장소의 default branch와 현재 원격 작업 브랜치를 먼저 확인한다.
- 수정·commit·push 직전에 원격 최신 상태를 다시 확인하고 다른 Agent/사용자의 최신 커밋을 덮어쓰지 않는다.
- force push로 최신 작업을 덮어쓰지 않는다.
- 오래된 diverged agent/실험 브랜치는 현재 작업선으로 간주하지 않는다.

## 2. 빌드·테스트 검증 우선순위

### 1순위 — subpc GitHub Actions self-hosted runner

- 해당 저장소의 self-hosted runner가 등록되어 있고 Online이면 우선 사용한다.
- 저장소의 실제 manifest, 기존 workflow, 프로젝트 문서에 정의된 test/build 명령을 사용한다.
- 실패하면 GitHub Actions의 job/step/log를 확인하고 원인을 수정한 뒤 다시 실행하여 결과를 확인한다.
- Windows 네이티브, Android 등 특정 플랫폼 검증은 가능한 경우 해당 플랫폼의 self-hosted runner 결과를 최종 근거로 사용한다.

### 2순위 — GitHub-hosted Actions fallback

다음과 같은 경우 GitHub-hosted runner를 fallback으로 사용할 수 있다.

- subpc가 꺼져 있거나 runner가 Offline인 경우
- subpc가 유지보수/장애/다른 작업으로 사용 불가능한 경우
- 필요한 플랫폼을 self-hosted 환경에서 제공할 수 없는 경우

GitHub-hosted Actions는 금지 대상이 아니다. 월간 사용 가능량이 남아 있고 적절한 hosted 환경이 있으면 fallback으로 사용할 수 있다.

self-hosted job의 `runs-on`이 자동으로 hosted runner로 전환되는 것은 아니므로, hosted fallback이 필요하면 기존 workflow를 확인한 뒤 명시적인 fallback 경로를 사용한다.

### 3순위 — 현재 Agent 환경에서 직접 test/build

self-hosted와 GitHub-hosted를 사용할 수 없거나 hosted 사용 가능량이 없는 경우, 현재 Agent 환경에서 가능한 test/build/lint/smoke를 직접 수행한다.

단, 현재 환경이 실제 대상 플랫폼을 재현하지 못하면 가능한 범위의 검증과 최종 플랫폼 검증을 구분한다.

### 4순위 — 사용자에게 직접 빌드·테스트 요청

위 세 방법으로도 필요한 최종 빌드/테스트를 확인할 수 없으면 성공으로 단정하지 않는다.

사용자에게 다음 취지로 명확히 요청한다.

> 현재 사용 가능한 실행 환경에서는 최종 빌드/테스트까지 확인하지 못했습니다. 가능하시면 직접 빌드/테스트해 보시고 결과를 알려주세요.

사용자가 결과나 오류 로그를 제공하면 그 결과를 기준으로 작업을 이어간다.

## 3. 성공 판정 기준

- 소스 검토나 정적 분석만으로 `빌드 성공`, `테스트 성공`, `검증 완료`라고 단정하지 않는다.
- 실제 명령의 정상 종료, 테스트 통과, 필요한 산출물 생성 등 실행 근거를 확인한다.
- runner Offline, quota, SDK/PATH, 서비스 장애 같은 인프라 실패와 프로젝트 코드 실패를 구분한다.
- 실패가 확인되면 가능한 경우 `로그 확인 → 수정 → 재실행 → 성공 확인`까지 닫힌 루프로 처리한다.

## 4. Artifact

- EXE, APK/AAB, ZIP 등 사용자가 받을 최종 산출물이 유용한 프로젝트는 성공 빌드 후 GitHub Actions artifact를 사용할 수 있다.
- artifact가 생성되면 필요 시 ChatGPT가 GitHub에서 가져와 사용자에게 전달할 수 있다.
- artifact 생성 자체는 테스트 통과를 대신하지 않는다.

## 5. GPU 작업

모델 학습이나 본격적인 GPU 검증은 기본 CI 대상으로 보지 않는다. 프로젝트에서 지정한 Colab, Kaggle, 전용 GPU 환경 등을 사용한다. CPU에서 가능한 preflight/gate/test는 위 우선순위에 따라 수행할 수 있다.

## 6. Workflow 운영

- self-hosted runner가 가용하면 기본 검증 경로로 사용한다.
- 기존 workflow가 있으면 불필요하게 중복 workflow를 만들기보다 현재 workflow를 우선 재사용·수정한다.
- 문서/정책만 바뀌어 실제 빌드가 불필요한 커밋은 필요 시 `[skip ci]`를 사용할 수 있다.
- 검증 결과를 보고할 때 어떤 경로를 사용했는지(`SELF_HOSTED`, `GITHUB_HOSTED`, `LOCAL`, `USER_REQUIRED`)와 실제 결과를 구분해서 기록한다.
