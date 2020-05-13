package mem

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestMemEventBus(t *testing.T) {
	s := NewMemEventBus()

	sub := s.Subscribe("test", func(ctx context.Context, data []byte) {
		fmt.Println(string(data))
	})

	for i := 0; i < 10; i++ {
		catch(s.Publish(context.Background(), "test", []byte(strconv.Itoa(i))))
	}

	time.Sleep(500 * time.Millisecond)
	sub.Unsubscribe()
}

func catch(err error) {}
