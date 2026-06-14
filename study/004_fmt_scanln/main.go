package main

import (
	"fmt"
)

func main() {
	var a, b int

	fmt.Print("숫자 2개 입력: ")

	// Scanln:
	// - 표준 입력에서 값을 공백(스페이스, 탭, 줄바꿈) 기준으로 나눠서 읽는다
	// - "한 줄 입력이 끝난 상태(Enter)"를 기준으로 입력을 확정하는 방식
	//
	// 동작 방식:
	// 1) 입력 스트림에서 공백 기준으로 토큰을 분리
	// 2) a, b 순서대로 값을 채움
	// 3) 줄 끝(개행)이 맞지 않거나 입력이 부족하면 에러 발생 가능
	//
	// 예:
	// - 정상: "1 2" + Enter → a=1, b=2
	// - 비정상: "1 Enter 2 Enter" → 흐름이 끊겨 에러 가능
	//
	// 특징:
	// - 입력을 "라인 단위"로 끝내야 한다는 제약이 있음
	// - 간단한 CLI 입력에는 편하지만, 유연성은 낮음

	_, err := fmt.Scanln(&a, &b)
	if err != nil {
		fmt.Println("입력 오류:", err)
		return
	}

	fmt.Println("a =", a)
	fmt.Println("b =", b)
}
