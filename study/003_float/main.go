package main

import (
	"fmt"
)

func main() {
	// 예: 가격 10.99와 20.50을 더하기
	price1 := 1099 // 10.99 * 100
	price2 := 2050 // 20.50 * 100

	total := price1 + price2                 // 3149
	fmt.Printf("%.2f", float64(total)/100.0) // 출력할 때만 실수로 변환: 31.49
}
