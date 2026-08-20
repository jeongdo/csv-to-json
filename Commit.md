# Commit Message

커밋 메시지는 작업 흐름을 나중에 복구할 수 있도록 작성한다.

## 커밋 단위

- 시작 전 상태 확인
- 완결된 이전 작업 먼저 커밋
- 의미 있는 작은 단위마다 커밋
- 목적이 다른 변경은 분리
- 사용자·다른 Agent 변경은 제외

## 형식

```text
<type>(phase-N): <summary>

Why:
Test:
Next:
```

## Phase

Phase를 사용하는 작업의 단위와 진행 규칙은 `Phase.md`를 따른다.

## 규칙

- `summary`는 무엇을 변경했는지 짧게 작성한다.
- `Why`, `Test`, `Next`는 필요한 경우에만 작성한다.
- Phase를 사용하는 프로젝트에서는 `<type>(phase-N)` 형식을 기본으로 사용한다.
- 다음 작업이 정해져 있으면 `Next`에 기록한다.
