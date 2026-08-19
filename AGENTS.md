# Agent Rules

## Build / Test 검증 정책 (최상위)

코드 변경 후 빌드·테스트 검증은 **루트 `BUILD.md`를 먼저 읽고 반드시 따른다.**

`BUILD.md`가 이 저장소의 self-hosted runner, GitHub-hosted Actions fallback, Agent 직접 검증, 사용자 직접 검증 요청에 대한 단일 기준이다.

과거 문서나 지침에 남아 있는 GitHub Actions 제한/금지 규칙과 충돌하면 **`BUILD.md`가 우선한다.**

또한 `main`을 고정으로 가정하지 말고, 작업 전에 default branch와 현재 원격 작업선을 확인하며 수정·push 직전 최신 상태를 다시 확인한다. 다른 Agent/사용자의 최신 커밋을 덮어쓰지 않는다.
