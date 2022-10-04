package stream_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/web/stream"
)

type Data struct {
	Msg string
}

func TestStream(t *testing.T) {
	service := stream.NewService[Data]()

	messages := []Data{
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

	counter := uint32(0)

	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(1000*time.Second))
	streams := make([]*mockStream[Data], 0)

	wg := sync.WaitGroup{}
	for i := 1; i < 10; i++ {
		st := service.AddStream(uint64(1), NewMockStream[Data](ctx, uint64(1), &counter))
		streams = append(streams, st.(*mockStream[Data]))
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := st.Run()
			fmt.Println(err)
		}()
		//for _, data := range messages {
		for j := 0; j < len(messages); j++ {
			//service.SendTo(&data, 1)
			service.SendTo(&messages[j], 1)
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
