package main

import "fmt"

// 1. 기본 함수
func add(a int, b int) int {
	return a + b
}

// 2. 다중 반환값 (Go의 상징)
// 에러를 두 번째 인자로 반환하는 패턴이 Go의 국룰이죠.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("0으로 나눌 수 없습니다")
	}
	return a / b, nil
}

// 3. 이름 지정 반환 (Named Return) & Naked Return
// 반환할 변수명을 미리 지정해두면, return만 써도 알아서 반환됩니다.
func getRectStats(width, height int) (area int, perimeter int) {
	area = width * height
	perimeter = 2 * (width + height)
	return // Naked return: area와 perimeter가 자동으로 반환됨
}

// 4. 가변 인자 (Variadic Functions)
// 인자의 개수가 정해지지 않았을 때 사용합니다.
func sumAll(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 5. [Go만의 특수 문법] defer (지연 실행)
// 함수가 종료되기 직전에 무조건 실행되는 코드를 예약합니다. (주로 자원 해제용)
func doSomething() {
	defer fmt.Println("함수 종료됨! (마지막에 실행)")
	fmt.Println("작업 중...")
}

func main() {
	// 1. 기본 함수 호출
	sum := add(2, 3)
	fmt.Printf("1. Add(2, 3) 결과: %d\n", sum)

	fmt.Println("--------------------------------")

	// 2. 다중 반환값 및 에러 처리 호출
	res, err := divide(10.0, 2.0)
	if err != nil {
		fmt.Printf("2. 에러 발생: %v\n", err)
	} else {
		fmt.Printf("2. Divide(10, 2) 결과: %.1f\n", res)
	}

	fmt.Println("--------------------------------")

	// 2. 다중 반환값 및 에러 처리 호출
	res1, err1 := divide(10.0, 0)
	if err1 != nil {
		fmt.Printf("2. 에러 발생: %v\n", err1)
	} else {
		fmt.Printf("2. Divide(10, 2) 결과: %.1f\n", res1)
	}

	fmt.Println("--------------------------------")

	// 3. 이름 지정 반환 & Naked Return 호출
	area, peri := getRectStats(100, 50)
	fmt.Printf("3. 사각형 - 넓이: %d, 둘레: %d\n", area, peri)

	fmt.Println("--------------------------------")

	// 4. 가변 인자 함수 호출
	total := sumAll(1, 2, 3, 4, 5)
	fmt.Printf("4. SumAll(1..5) 결과: %d\n", total)

	fmt.Println("--------------------------------")

	// 5. defer 테스트
	fmt.Println("5. Defer 테스트 시작:")
	doSomething()
}
