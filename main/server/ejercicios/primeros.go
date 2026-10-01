package ejercicios

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type fnInt func(n int) int

func (f *fnInt) doble(n int, ch chan<- int) (int, error) {
	fn := *f

	if fn == nil {
		return 0, errors.New("Función Vacía!...")
	}

	time.Sleep(time.Duration(n) * time.Second)

	r := fn(n) * 2
	ch <- r

	return r, nil
}

func Init() {

	var fn fnInt = func(n int) int { return n % 5 }
	var wg sync.WaitGroup

	ch := make(chan int)

	for i := range 10 {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			_, err := fn.doble(val, ch)
			if err != nil {
				fmt.Println("Error: " + err.Error())
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		fmt.Println(r)
	}

}
