package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

var (
	tasks    int
	errIndex int
	silent   bool

	sleep = 10 * time.Millisecond
)

func init() {
	flag.IntVar(&tasks, "tasks", 10, "number of tasks to do")
	flag.IntVar(&errIndex, "error-index", 0, "the index to error on")
	flag.BoolVar(&silent, "silent", false, "be silent")

	rand.Seed(time.Now().UTC().UnixNano())
}

func log(format string, args ...interface{}) {
	if !silent {
		fmt.Printf(format, args...)
	}
}

func work(i int) (int, error) {
	d := time.Duration(float64(sleep) * rand.Float64())
	log("sleeping %d: %s\n", i, d)
	time.Sleep(d)
	if i == errIndex {
		return 0, fmt.Errorf("bang %d!", i)
	}
	return i, nil
}

type res struct {
	i   int
	err error
}

func main() {
	flag.Parse()

	fmt.Printf("context test\n  number: %d\n  error:  %d\n  silent: %t\n\n", tasks, errIndex, silent)

	ch := make(chan res)
	defer close(ch)

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < tasks; i++ {
		go func(ctx context.Context, i int) {
			v, err := work(i)
			select {
			case <-ctx.Done():
				log("context closed: %d\n", i)
			default:
				log("sending: %d, %d, %v\n", i, v, err)
				ch <- res{i: v, err: err}
			}
		}(ctx, i)
	}

	var ints []int
	var err error

	for res := range ch {
		if res.err != nil {
			cancel()
			err = res.err
			break
		}
		ints = append(ints, res.i)
		if len(ints) == tasks {
			break
		}
	}

	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	log("result: %v\n", ints)
	fmt.Printf("got %d results\n", len(ints))

	// give all tasks time to finish
	time.Sleep(sleep)
}
