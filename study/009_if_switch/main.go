package main

import (
	"fmt"
)

func main() {

	fmt.Println("======================================")
	fmt.Println("1. 기본 if")
	fmt.Println("======================================")

	basicIf()

	fmt.Println("\n======================================")
	fmt.Println("2. if - else")
	fmt.Println("======================================")

	ifElse()

	fmt.Println("\n======================================")
	fmt.Println("3. if - else if")
	fmt.Println("======================================")

	elseIf()

	fmt.Println("\n======================================")
	fmt.Println("4. if 초기화문")
	fmt.Println("======================================")

	ifInitializer()

	fmt.Println("\n======================================")
	fmt.Println("5. 중첩 if")
	fmt.Println("======================================")

	nestedIf()

	fmt.Println("\n======================================")
	fmt.Println("6. 기본 switch")
	fmt.Println("======================================")

	basicSwitch()

	fmt.Println("\n======================================")
	fmt.Println("7. 여러 case")
	fmt.Println("======================================")

	multiCaseSwitch()

	fmt.Println("\n======================================")
	fmt.Println("8. switch true")
	fmt.Println("======================================")

	switchWithoutValue()

	fmt.Println("\n======================================")
	fmt.Println("9. fallthrough")
	fmt.Println("======================================")

	fallthroughExample()

	fmt.Println("\n======================================")
	fmt.Println("10. switch 초기화문")
	fmt.Println("======================================")

	switchInitializer()

	fmt.Println("\n======================================")
	fmt.Println("11. type switch")
	fmt.Println("======================================")

	typeSwitchExample()
}

func basicIf() {

	age := 20

	if age >= 19 {
		fmt.Println("성인")
	}
}

func ifElse() {

	score := 50

	if score >= 60 {
		fmt.Println("합격")
	} else {
		fmt.Println("불합격")
	}
}

func elseIf() {

	score := 85

	if score >= 90 {
		fmt.Println("A")
	} else if score >= 80 {
		fmt.Println("B")
	} else if score >= 70 {
		fmt.Println("C")
	} else {
		fmt.Println("F")
	}
}

func ifInitializer() {

	/*
		if age := 25; age >= 20 {

		}
		age := 25 → 변수 선언
		age >= 20 → 조건 체크

		동시에 한 줄에서 처리하는 겁니다.
		실제로 자주 쓰는 패턴은 이런 거예요:
		if err := someFunc(); err != nil {
			// 에러 처리
		}
	*/

	// if 안에서 변수 선언
	if age := 25; age >= 20 {
		fmt.Println("20대 이상")
	}

	// age는 여기서 사용 불가
	// fmt.Println(age) // 컴파일 에러
}

func nestedIf() {

	age := 25
	hasLicense := true

	if age >= 20 {

		if hasLicense {
			fmt.Println("운전 가능")
		} else {
			fmt.Println("면허 없음")
		}
	}
}

func basicSwitch() {

	day := "MON"

	switch day {

	case "MON":
		fmt.Println("월요일")

	case "TUE":
		fmt.Println("화요일")

	case "WED":
		fmt.Println("수요일")

	default:
		fmt.Println("기타")
	}
}

func multiCaseSwitch() {

	day := "SAT"

	switch day {

	case "SAT", "SUN":
		fmt.Println("주말")

	case "MON", "TUE", "WED", "THU", "FRI":
		fmt.Println("평일")

	default:
		fmt.Println("알 수 없음")
	}
}

func switchWithoutValue() {

	score := 85

	// switch true 와 동일
	switch {

	case score >= 90:
		fmt.Println("A")

	case score >= 80:
		fmt.Println("B")

	case score >= 70:
		fmt.Println("C")

	default:
		fmt.Println("F")
	}
}

func fallthroughExample() {

	level := 1

	switch level {

	case 1:
		fmt.Println("LEVEL 1")

		// 다음 case 강제 실행
		fallthrough

	case 2:
		fmt.Println("LEVEL 2")

	case 3:
		fmt.Println("LEVEL 3")
	}
}

func switchInitializer() {

	switch age := 30; {

	case age >= 60:
		fmt.Println("노년")

	case age >= 20:
		fmt.Println("성인")

	default:
		fmt.Println("미성년")
	}
}

func typeSwitchExample() {

	printType(100)
	printType("hello")
	printType(true)
	printType(3.14)
}

func printType(v interface{}) {

	switch value := v.(type) {

	case int:
		fmt.Println("int 타입 =", value)

	case string:
		fmt.Println("string 타입 =", value)

	case bool:
		fmt.Println("bool 타입 =", value)

	case float64:
		fmt.Println("float64 타입 =", value)

	default:
		fmt.Printf("알 수 없는 타입: %T\n", value)
	}
}
