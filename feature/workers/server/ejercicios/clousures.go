package ejercicios

import "fmt"

func account() (func(int), func() int) {

	amount := 1000

	debit := func(x int) {
		amount -= x
	}

	getAmount := func() int {
		return amount
	}

	return debit, getAmount
}

func InitClousure() {
	debit, getAmount := account()

	fmt.Println(getAmount())
	debit(100)
	fmt.Println(getAmount())

}
