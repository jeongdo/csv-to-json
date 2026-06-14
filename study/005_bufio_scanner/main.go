package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	/*
		버퍼 없으면: 매 글자마다 OS 호출 / 속도 느림 (I/O 비용 큼)
		버퍼 있으면: OS → 한 번에 8KB 정도 읽음 → 메모리 저장 → 거기서 잘라씀
	*/
	
	// bufio.Reader:
	// - 입력(os.Stdin)을 버퍼로 감싸서 "한 줄 단위로 직접 제어"할 수 있게 해준다
	// - Scanln처럼 자동 파싱이 아니라 개발자가 직접 문자열을 처리한다
	stdin := bufio.NewReader(os.Stdin)

	fmt.Print("숫자 2개를 공백으로 구분해 입력하세요: ")

	// ReadString('\n'):
	// - 엔터('\n')가 나올 때까지 입력 전체를 문자열로 읽는다
	// - 즉, "한 줄 통째로 입력 받는 방식"
	line, err := stdin.ReadString('\n')
	if err != nil {
		fmt.Println("입력 에러:", err)
		return
	}

	// TrimSpace:
	// - 문자열 양 끝의 공백, 개행(\n, \r\n)을 제거한다
	// - 입력 정리 단계 (파싱 전에 필수로 자주 사용됨)
	line = strings.TrimSpace(line)

	// Sscanf:
	// - 문자열을 특정 포맷에 맞춰 파싱한다
	// - 여기서는 "정수 2개"를 문자열에서 추출
	var a, b int
	_, err = fmt.Sscanf(line, "%d %d", &a, &b)
	if err != nil {
		fmt.Println("입력 형식 오류:", err)
		return
	}

	fmt.Println("a =", a)
	fmt.Println("b =", b)
}
