package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/dyrkin/rezerwacje-duw-go/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestUniquiness(c *C) {
	queue := New()
	userData := []*config.Row{&config.Row{Name: "hello", Value: "world"}}
	entity := &config.Entity{Name: "name", ShortName: "short", Queue: "10", ID: "100"}
	reservation := &Reservation{Entity: entity, Date: "2017-07-23", Term: "13:20", UserData: &userData}
	for i := 0; i < 3; i++ {
		queue.Push(reservation)
	}
	c.Assert(queue.Len(), Equals, 1)
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-23", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
	queue.Push(reservation)
	c.Assert(queue.Len(), Equals, 1)
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-23", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
}

func (s *MySuite) TestOrdering(c *C) {
	queue := New()
	userData := []*config.Row{&config.Row{Name: "hello", Value: "world"}}
	entity := &config.Entity{Name: "name", ShortName: "short", Queue: "10", ID: "100"}
	for i := 0; i < 3; i++ {
		reservation := &Reservation{Entity: entity, Date: fmt.Sprintf("2017-07-2%d", i), Term: "13:20", UserData: &userData}
		queue.Push(reservation)
	}
	c.Assert(queue.Len(), Equals, 3)
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-22", Term: "13:20", UserData: &userData})
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-21", Term: "13:20", UserData: &userData})
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-20", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
}

func (s *MySuite) TestTake(c *C) {
	queue := New()
	userData := []*config.Row{&config.Row{Name: "hello", Value: "world"}}
	entity := &config.Entity{Name: "name", ShortName: "short", Queue: "10", ID: "100"}

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(100 * time.Millisecond)
			reservation := &Reservation{Entity: entity, Date: fmt.Sprintf("2017-07-2%d", i), Term: "13:20", UserData: &userData}
			queue.Push(reservation)
		}
	}()

	c.Assert(queue.Take(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-20", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
	c.Assert(queue.Take(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-21", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
	c.Assert(queue.Take(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-22", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
}

func (s *MySuite) TestLimit(c *C) {
	queue := NewWithLimit(3)
	userData := []*config.Row{&config.Row{Name: "hello", Value: "world"}}
	entity := &config.Entity{Name: "name", ShortName: "short", Queue: "10", ID: "100"}
	for i := 0; i < 5; i++ {
		reservation := &Reservation{Entity: entity, Date: fmt.Sprintf("2017-07-2%d", i), Term: "13:20", UserData: &userData}
		queue.Push(reservation)
	}
	c.Assert(queue.Len(), Equals, 3)
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-24", Term: "13:20", UserData: &userData})
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-23", Term: "13:20", UserData: &userData})
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-22", Term: "13:20", UserData: &userData})
	var r *Reservation
	c.Assert(queue.Pop(), DeepEquals, r)
	c.Assert(queue.Len(), Equals, 0)
	queue.Push(&Reservation{Entity: entity, Date: "2017-07-22", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 1)
	c.Assert(queue.Pop(), DeepEquals, &Reservation{Entity: entity, Date: "2017-07-22", Term: "13:20", UserData: &userData})
	c.Assert(queue.Len(), Equals, 0)
}
