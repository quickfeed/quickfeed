package stream_test

import (
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/web/stream"
)

type Data struct {
	Msg string
}

var messages = []*Data{
	{Msg: "Hello"},
	{Msg: "World"},
	{Msg: "Foo"},
	{Msg: "Bar"},
	{Msg: "Gandalf"},
	{Msg: "Frodo"},
	{Msg: "Bilbo"},
	{Msg: "Radagast"},
	{Msg: "Sauron"},
	{Msg: "Gollum"},
}

func TestStream(t *testing.T) {
	service := stream.NewService[uint64, Data]()
	streams := make([]*qtest.MockStream[Data], 0)

	var counter uint32
	wg := sync.WaitGroup{}
	for i := 1; i < 10; i++ {
		stream := qtest.NewMockStream[Data](t)
		stream.SetCounter(&counter)
		service.Add(stream, 1)
		streams = append(streams, stream)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := stream.Run()
			t.Log(err)
		}()
		for _, data := range messages {
			data := data
			service.SendTo(data, 1)
		}
	}

	service.CloseBy(1)
	wg.Wait()

	// A total of 90 messages should have been sent.
	if counter != 90 {
		t.Errorf("expected 90, got %d", counter)
	}

	for _, s := range streams {
		msgMsp := make(map[string]int)
		for _, data := range s.Messages {
			msgMsp[data.Msg]++
		}
		t.Log(msgMsp)
		if len(s.Messages) != 10 {
			t.Errorf("expected 10 messages, got %d", len(s.Messages))
		}
		if diff := cmp.Diff(messages, s.Messages); diff != "" {
			t.Errorf("expected %v, got %v: %s", messages, s.Messages, diff)
		}
	}
}

// TestStreamClose tries to send messages to a stream that is closing.
// This test should be run with the -race flag, e.g.,:
// % go test -v -race -run TestStreamClose -test.count 10
func TestStreamClose(t *testing.T) {
	service := stream.NewService[uint64, Data]()

	stream := qtest.NewMockStream[Data](t)
	service.Add(stream, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 1_000_000; i++ {
			stream.Send(messages[i%len(messages)])
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = stream.Run()
	}()
	stream.Close()
	wg.Wait()
}
