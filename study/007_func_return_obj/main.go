package main

import "fmt"

// Go는 구조체로 데이터를 표현합니다.
type Developer struct {
	Name string
	Age  int
}

// [Go의 관습] New + 구조체이름으로 생성자 역할을 하는 함수를 만듭니다.
// 포인터(*Developer)를 반환하여 메모리 주소를 넘겨주는 것이 일반적입니다.
/*
1. 탈출 분석(Escape Analysis)이란?
	Go 컴파일러는 코드를 컴파일할 때, 함수 내부에서 생성된 변수가 "함수 외부로 탈출(Escape)해서 계속 사용되는지"를 정밀하게 추적합니다.
	탈출하지 않는 경우: 변수가 함수 내부에서만 쓰이고 끝난다면, 가장 빠르고 가벼운 스택(Stack) 메모리에 할당합니다. 함수가 끝나는 순간 흔적도 없이 사라집니다.
	탈출하는 경우: 위 코드처럼 &dev를 통해 주소값을 함수 밖으로 던지면, 컴파일러는 "이 변수는 함수가 끝나도 밖에서 누군가 쓰겠구나(탈출했구나)!"라고 판단합니다. 그리고 개발자가 시키지 않아도 알아서 이 객체를 힙(Heap) 메모리에 할당하도록 설계를 변경합니다.
2. 이 주석이 베테랑에게 주는 의미
	결국 이 주석은 "Go에서는 포인터를 리턴해도 메모리 오염이나 버그가 발생하지 않으니 안심하라"는 뜻입니다.
	C/C++과의 차이: C 언어에서는 힙에 올리려면 반드시 malloc을 쓰고 나중에 free를 해야 했습니다. 하지만 Go에서는 & 하나만 붙여서 밖으로 던지면 컴파일러가 알아서 힙으로 보내고, 나중에 가비지 컬렉터(GC)가 수거해 갑니다.
	포인터 연산의 안전성: 함수가 종료되어 NewDeveloper 스택 프레임은 파괴되지만, dev 객체 자체는 안전한 힙 공간에 살아남아 이 함수를 호출한 메인 로직에게 주소값을 전달하게 됩니다.
*/

func NewDeveloper(name string, age int) *Developer {
	dev := Developer{Name: name, Age: age}
	// "함수 안에서 지역 변수로 선언한 객체의 주소(&dev)를 리턴?
	// 함수 끝나면 스택(Stack) 프레임이 날아가서 댕글링 포인터(Dangling Pointer)가 될 텐데?!"
	return &dev // 탈출 분석 덕분에 함수가 끝나도 이 객체는 힙에 살아남습니다.
}

// defer의 활용 예시
func processResource() {
	fmt.Println("1. 자원(파일/락)을 획득했습니다.")

	// defer는 함수의 멱살을 잡고 있다가, 함수가 끝나기 직전(return 이후)에 무조건 실행됩니다.
	defer fmt.Println("3. [defer] 자원을 안전하게 해제했습니다. (Go 방식)")

	fmt.Println("2. 자원을 가지고 핵심 로직을 수행 중입니다...")
}

func main() {
	d := NewDeveloper("허정도", 49)
	fmt.Printf("Go Developer 생성: %+v\n", d)
	processResource()
}
