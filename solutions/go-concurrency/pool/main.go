package main

import (
	"fmt"
	"sync"
)

var numbers = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

const numWorkers = 3

func main() {

	input := make(chan int)
	output := make(chan int)
	var wg sync.WaitGroup

	// pool feeder
	go func() {
		for _, i := range numbers {
			input <- i
		}
		close(input)
	}()

	// pool of workers
	for n := range numWorkers {
		wg.Go(func() {
			fmt.Printf("#%d started\n", n)

			// process until end of input
			for number := range input {
				fmt.Printf("#%d processing: %d\n", n, number)
				output <- number * number
			}
			fmt.Printf("#%d finished\n", n)
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
