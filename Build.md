# Build / Test Environment

| 우선순위 | 환경 | 설명 | 사용 시점 |
|---:|---|---|---|
| 1 | **LOCAL** | 저장소 내 로컬 빌드/테스트 환경 (Go, Cargo, npm, Gradle, pytest 등) | 기본 검증. 실제 빌드·테스트가 정상 통과하면 일반적인 검증은 완료로 본다. |
| 2 | **SELF_HOSTED** | GitHub Actions self-hosted runner (subpc 등) | LOCAL에서 대상 OS/SDK 재현이 어렵거나, 도구 부족·장시간·고자원 작업 등 추가 검증이 필요한 경우 |
| 3 | **GITHUB_HOSTED** | GitHub-hosted runner (`ubuntu-latest`, `windows-latest` 등) | SELF_HOSTED를 사용할 수 없거나 깨끗한 CI 환경에서 별도 재현 확인이 필요한 경우 |
| 4 | **USER_REQUIRED** | 사용자에게 직접 빌드/테스트 요청 | 자동화된 환경에서 최종 검증이 불가능한 경우 |
