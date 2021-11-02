package http

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestResponseRequest_next(t *testing.T) {
	var (
		until = time.Now().Add(300 * time.Millisecond)
		rs    = ResponseSequence{
			Responses: []*Response{
				{
					Body: []byte("1"),
					Repeat: Repeat{
						Until: until,
					},
				},
				{
					Body: []byte("2"),
					Repeat: Repeat{
						Count: 3,
					},
				},
				{
					Body: []byte("3"),
					Repeat: Repeat{
						For: 300 * time.Millisecond,
					},
				},
			},
		}

		is = is.New(t)
	)

	for time.Now().Before(until) {
		is.Equal(rs._next().Body, []byte("1"))
		time.Sleep(10 * time.Millisecond)
	}

	for i := 0; i < 3; i++ {
		is.Equal(rs._next().Body, []byte("2"))
	}

	until = time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(until) {
		is.Equal(rs._next().Body, []byte("3"))
		time.Sleep(10 * time.Millisecond)
	}

	is.Equal(rs._next(), nil)
}
