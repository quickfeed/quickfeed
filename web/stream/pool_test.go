package stream_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/web/stream"
)

type data struct {
	Msg string
}

func TestPool(t *testing.T) {
	pool := stream.NewPool[data]()

	// Add a channel to the pool.
	ch := pool.Add(1)
	ch2 := pool.Add(2)
	ch3 := pool.Add(3)

	// Send data to all channels in the pool.
	go pool.Send(&data{Msg: "hello"})

	// Read data from the channels.
	for _, ch := range []chan *data{ch, ch2, ch3} {
		d := <-ch
		if d.Msg != "hello" {
			t.Errorf("got %q, want %q", d.Msg, "hello")
		}
	}

	go pool.SendTo(2, &data{Msg: "hello 2"})
	d := <-ch2
	if d.Msg != "hello 2" {
		t.Errorf("got %q, want %q", d.Msg, "hello 2")
	}

	go pool.SendTo(3, &data{Msg: "hello 3"})
	d = <-ch3
	if d.Msg != "hello 3" {
		t.Errorf("got %q, want %q", d.Msg, "hello 3")
	}

	// Remove a channel from the pool.
	pool.Remove(1)
	if _, ok := <-ch; ok {
		t.Errorf("channel should be closed")
	}

	// Send data to all channels in the pool.
	go pool.Send(&data{Msg: "hello again"})
	for _, ch := range []chan *data{ch2, ch3} {
		d := <-ch
		if d.Msg != "hello again" {
			t.Errorf("got %q, want %q", d.Msg, "hello again")
		}
	}

	// Close all channels in the pool.
	pool.Close()
	for _, ch := range []chan *data{ch2, ch3} {
		if _, ok := <-ch; ok {
			t.Errorf("channel should be closed")
		}
	}

	// Add two channels to the pool with the same ID.
	// The first channel should be closed and replaced.
	ch4 := pool.Add(3)
	_ = pool.Add(3)
	// ch4 should be closed.
	if _, ok := <-ch4; ok {
		t.Errorf("channel should be closed")
	}
}
