package main

import (
	"fmt"
	"sync"
)

var numbers = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

const numWorkers = 3

// worker's clean function
func worker(number int) int {
	return number * number
}

func main() {

	// feeder
	input := make(chan int)
	go func() {
		for _, i := range numbers {
			input <- i
		}
		close(input)
	}()

	// pool of "clean" workers
	output := make(chan int)
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for number := range input {
				output <- worker(number)
			}
		})
	}

	// output closer
	go func() {
		wg.Wait()
		close(output)
	}()

	// collect results
	for result := range output {
		fmt.Println(result)
	}
}
