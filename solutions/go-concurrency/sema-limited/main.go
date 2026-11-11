package main

import (
	"fmt"
)

var numbers = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

const numWorkers = 3

func main() {

	output := make(chan int)
	sema := make(chan struct{}, numWorkers)

	// run in separate goroutine to avoid deadlock
	// - someone needs to read output channel
	go func() {
		for n, number := range numbers {
			// book semaphore to start
			sema <- struct{}{}
			go func() {
				fmt.Printf("#%d processing: %d\n", n, number)
				output <- number * number
				<-sema
				fmt.Printf("#%d finished\n", n)
			}()
		}
	}()

	// collect results - we know exact number of them
	for range len(numbers) {
		result := <-output
		fmt.Println("## result:", result)
	}
}
