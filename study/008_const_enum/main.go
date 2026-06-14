package main

import "fmt"

// ============================================================
// Go의 enum 패턴 정리
//
// Go에는 Java, Rust처럼 enum 키워드가 없습니다.
// 대신 const + iota, 메서드, 인터페이스를 조합해서 enum을 흉내냅니다.
// ============================================================

// ============================================================
// 1. 기본형: const + iota
//
// 가장 단순한 형태. 상수에 순차적인 숫자를 매핑합니다.
//
// 포인트:
//   - 단순 int 대신 "type Status int"로 타입을 분리하는 게 핵심.
//   - Status와 일반 int는 서로 다른 타입이 됩니다.
//   - Status를 받는 함수에 int를 넘기면 컴파일 에러 → 타입 안전성 확보.
// ============================================================

/*

	type Status int  // 새로운 타입 (distinct type)
	type Status = int // 그냥 별칭 (alias) - = 있으면 완전히 같은 타입 취급

	type Status int

	func printStatus(s Status) {
		fmt.Println(s)
	}

	func main() {
		printStatus(1)        // 컴파일 에러! int는 Status가 아님
		printStatus(Status(1)) // OK - 명시적 변환 필요
	}

*/

type Status int

const (
	StatusIdle    Status = iota // 0 - iota는 const 블록 안에서 0부터 자동으로 1씩 증가
	StatusRunning               // 1
	StatusError                 // 2
)

// ============================================================
// 2. 확장형: 메서드를 포함한 enum 구조
//
// Go는 어떤 타입에도 메서드를 붙일 수 있습니다.
// 상태값에 관련 로직을 캡슐화할 수 있습니다.
// ============================================================

type OrderStatus int

const (
	Pending   OrderStatus = iota // 결제 대기
	Shipped                      // 배송 중
	Delivered                    // 배송 완료
)

// String() 메서드 = Stringer 패턴
// fmt.Println, fmt.Printf("%s") 등에서 자동으로 호출됨.
// 숫자 대신 사람이 읽을 수 있는 이름이 출력되어 디버깅/로깅이 편해짐.
func (os OrderStatus) String() string {
	switch os {
	case Pending:
		return "결제 대기 중"
	case Shipped:
		return "배송 중"
	case Delivered:
		return "배송 완료"
	default:
		return "알 수 없음"
	}
}

// Next() = 상태 전이 로직
// 상태에 관련된 로직을 메서드로 묶는 것이 Go식 캡슐화.
// 외부에서 if/switch로 상태를 분기하는 대신, 타입 자신이 다음 상태를 알고 있음.
func (os OrderStatus) Next() OrderStatus {
	switch os {
	case Pending:
		return Shipped
	case Shipped:
		return Delivered
	default:
		return os // Delivered는 더 이상 전이 없음 → 자기 자신 반환
	}
}

// ============================================================
// 3. 고도화: 인터페이스를 활용한 데이터 enum
//
// [왜 필요한가?]
// Rust/Swift의 enum은 variant마다 다른 데이터를 품을 수 있습니다:
//
//   // Rust
//   enum Action {
//       Stop,
//       Move { x: i32, y: i32 },  // 데이터가 붙은 variant
//   }
//
// Go의 const는 그냥 숫자라서 데이터를 붙이는 게 불가능합니다:
//
//   const (
//       Stop = iota
//       Move  // x, y를 어떻게 붙여? → 불가능
//   )
//
// 그래서 Go에서는 인터페이스 + 구조체로 이를 흉내냅니다.
// "데이터가 붙은 enum variant 하나 = 구조체 하나" 로 대응시키는 패턴.
//
// [Rust vs Go 비교]
//   항목                          Rust                          Go
//   enum 선언         enum Action { Stop, Move{x,y} }    인터페이스 + 구조체들
//   분기 처리          match action { ... } 한 블록        인터페이스 메서드 호출
//   미처리 케이스       컴파일러가 강제로 체크                런타임에서 확인
//   코드량             적음                               타입마다 구조체 + 메서드 필요
//
// 결론: Go의 3번 패턴은 언어 차원에서 지원하지 않는 걸
//       인터페이스로 억지로 구현한 것입니다.
//       Rust는 이걸 enum + match 한 방에 처리합니다.
// ============================================================

// Action 인터페이스 = enum 타입 자체.
// 이 인터페이스를 구현한 구조체들이 각각의 variant가 됨.
type Action interface {
	Execute()
}

// variant 1: 데이터 없는 상태 → 빈 구조체
type StopAction struct{}

func (s StopAction) Execute() {
	fmt.Println("멈춤!")
}

// variant 2: 데이터가 있는 상태 → 필드를 가진 구조체
// Rust의 Move { x, y }에 해당
type MoveAction struct {
	X, Y int
}

func (m MoveAction) Execute() {
	fmt.Printf("좌표 (%d, %d)로 이동!\n", m.X, m.Y)
}

/*
	class MoveAction implements Action {  // ← 자바 implements 써야 함
    	public void execute() { ... }
	}

	==================================================================

	type Action interface {
		Execute()
	}

	type MoveAction struct { X, Y int }

	func (m MoveAction) Execute() { ... }  // ← Go 이것만으로 Action 구현 완료

*/

// ============================================================
// main
// ============================================================

func main() {
	// --- 1. 기본형 ---
	fmt.Println("=== 1. 기본형 ===")
	var s Status = StatusRunning
	fmt.Printf("현재 상태: %d\n\n", s) // 출력: 1 (숫자 그대로 출력)

	// --- 2. 확장형 ---
	fmt.Println("=== 2. 확장형 ===")
	state := Pending
	fmt.Println("시작:", state) // String() 자동 호출 → "결제 대기 중"
	state = state.Next()
	fmt.Println("다음:", state) // "배송 중"
	state = state.Next()
	fmt.Println("다음:", state) // "배송 완료"
	fmt.Println()

	// --- 3. 고도화 ---
	fmt.Println("=== 3. 고도화 ===")
	// 인터페이스 타입으로 묶어서 하나의 슬라이스로 관리 가능
	// → Rust에서 같은 enum 타입으로 묶는 것과 동일한 효과
	actions := []Action{
		MoveAction{X: 10, Y: 20},
		StopAction{},
	}
	for _, a := range actions {
		a.Execute() // 실제 타입에 따라 알맞은 메서드가 호출됨 (다형성)
	}
}

// ============================================================
// 아키텍처 팁 정리
//
//  1. 타입 안전성:   int 대신 "type Status int"로 선언
//                   → 컴파일 타임 오류 검출
//
//  2. Stringer 패턴: String() 메서드 구현
//                   → 로그/디버그 시 숫자 대신 이름 출력
//
//  3. 메서드 캡슐화: 상태 로직을 메서드로 결합
//                   → 외부 분기문 제거, Go식 객체지향
// ============================================================
