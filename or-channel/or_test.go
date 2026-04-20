package or

import (
	"fmt"
	"testing"
	"time"
)

func signal(after time.Duration) <-chan interface{} {
	done := make(chan interface{})

	go func() {
		defer close(done)
		time.Sleep(after)
	}()

	return done
}

func TestOrClosesWhenFirstChannelCloses(t *testing.T) {
	start := time.Now()

	<-Or(
		signal(2*time.Hour),
		signal(5*time.Minute),
		signal(50*time.Millisecond),
		signal(1*time.Hour),
		signal(1*time.Minute),
	)

	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("or closed too late: %v", elapsed)
	}
}

func TestOrWithNoChannelsReturnsNil(t *testing.T) {
	if ch := Or(); ch != nil {
		t.Fatal("expected nil channel when no channels provided")
	}
}

func TestOrWithSingleChannelReturnsSameChannel(t *testing.T) {
	done := make(chan interface{})

	if got := Or(done); got != done {
		t.Fatal("expected the same channel to be returned for a single input")
	}
}

func ExampleOr() {
	sig := func(after time.Duration) <-chan interface{} {
		done := make(chan interface{})

		go func() {
			defer close(done)
			time.Sleep(after)
		}()

		return done
	}

	<-Or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(10*time.Millisecond),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Println("done")
	// Output: done
}
