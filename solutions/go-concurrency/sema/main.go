package main

import (
	"fmt"
	"sync"
)

var numbers = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

const numWorkers = 3

func main() {

	output := make(chan int)
	sema := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup

	for n, number := range numbers {
		wg.Go(func() {
			// book semaphore to start
			sema <- struct{}{}
			fmt.Printf("#%d started\n", n)

			fmt.Printf("#%d processing: %d\n", n, number)
			output <- number * number

			fmt.Printf("#%d finished\n", n)
			<-sema
		})
	}

	// output closer
	go func() {
		wg.Wait()
		close(output)
	}()

	// collect results
	for result := range output {
		fmt.Println("## result:", result)
	}
}
