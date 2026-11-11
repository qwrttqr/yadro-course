package main

import (
	"fmt"
	"sync"
)

const numWorkers = 3

// pool feeder
func feeder(numbers []int) <-chan int {
	input := make(chan int)
	go func() {
		for _, i := range numbers {
			input <- i
		}
		close(input)
	}()
	return input
}

// pool of "clean" workers
func fanOut(input <-chan int) []<-chan int {
	var workerChannels []<-chan int
	for range numWorkers {
		workerChannels = append(workerChannels, worker(input))
	}
	return workerChannels
}

// worker goroutine
func worker(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		for number := range in {
			out <- number * number
		}
		close(out)
	}()

	return out
}

// fan in - channel merging goroutine
func fanIn(in []<-chan int) <-chan int {

	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range in {
		wg.Go(func() {
			for number := range ch {
				out <- number
			}
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	// feeder -> fanout / pool -> fanin / collector -> printer
	for result := range fanIn(fanOut(feeder(numbers))) {
		fmt.Println(result)
	}
}
