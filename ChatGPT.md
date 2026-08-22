# ChatGPT

## Execution Time Limit

Chat 실행 환경에서는 하나의 실행 작업을 약 30초 이내의 안전한 단위로 나눈다.

장시간 작업은 가능한 경우 상태를 파일 또는 checkpoint로 저장한 뒤 실행을 종료하고, 다음 실행에서 이어서 수행한다.

재개 가능한 작업은 이 방식을 반복하여 장시간 수행할 수 있다.

이 규칙은 Chat 실행 환경에만 적용하며, Work 및 Codex 등 다른 실행 환경에는 적용하지 않는다.

## External Waiting

외부 작업 결과를 기다리기 위해 짧은 대기와 반복 확인을 사용할 수 있다. (GitHub 등)

## Plugins

적극 활용 플러그인:
- Google Drive
- Dropbox
