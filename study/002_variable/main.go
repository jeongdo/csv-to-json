package main

import (
	"fmt"
)

func main() {
	const Pi = 3.14159
	const AppName = "MySystem"
	
	var a int = 10 // a가 10 이런 뜻
	var msg string = "hello"

	// var a = 10         // 컴파일러가 10을 보고 자동으로 int로 결정
	// var msg = "hello"  // 컴파일러가 "hello"를 보고 string으로 결정

	// a := 10        // var 키워드도 생략, 타입도 생략
	// msg := "hello" // Go 개발자들은 거의 90% 이상 이렇게 씁니다.

	b := 10
	a = 20
	msg = "hello2"
	fmt.Println(a, msg, b)
}
