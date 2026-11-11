package main

import (
	"fmt"
	"sync"
)

var numbers = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

const numWorkers = 3

func worker(wg *sync.WaitGroup, in <-chan int, out chan<- int) {
	// need wg.Done? close any channels???

	defer wg.Done()
	for number := range in {
		out <- number * number
	}
}

func main() {

	output := make(chan int)
	input := make(chan int)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// pool of "dirty" workers
	for range numWorkers {
		go worker(&wg, input, output)
	}

	// pool feeder
	go func() {
		for _, i := range numbers {
			input <- i
		}
		close(input)
	}()

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
