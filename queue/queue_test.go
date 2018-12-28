package queue

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestOrdering1(c *C) {
	queue := New()
	switches := [3]byte{}
	for i := range switches {
		fn := func(ind int) func() {
			return func() { switches[ind] = 1 }
		}(i)
		queue.Push(fn)
	}
	c.Assert(queue.Len(), Equals, 3)
	queue.Pop()()
	c.Assert(switches, DeepEquals, [3]byte{0, 0, 1})
	c.Assert(queue.Len(), Equals, 2)
}

func (s *MySuite) TestOrdering2(c *C) {
	queue := New()
	switches := [4]byte{}
	for i := range switches {
		fn := func(ind int) func() {
			return func() { switches[ind] = 1 }
		}(i)
		queue.Push(fn)
	}
	c.Assert(queue.Len(), Equals, 4)
	queue.Pop()()
	queue.Pop()
	queue.Pop()()
	queue.Pop()
	c.Assert(switches, DeepEquals, [4]byte{0, 1, 0, 1})
	c.Assert(queue.Len(), Equals, 0)
}

func (s *MySuite) TestTake(c *C) {
	queue := New()
	switches := [4]byte{}
	go func() {
		for i := range switches {
			time.Sleep(100 * time.Millisecond)
			fn := func(ind int) func() {
				return func() { switches[ind] = 1 }
			}(i)
			queue.Push(fn)
		}
	}()
	queue.Take()()
	c.Assert(queue.Len(), Equals, 0)
	queue.Take()
	c.Assert(queue.Len(), Equals, 0)
	queue.Take()()
	c.Assert(queue.Len(), Equals, 0)
	queue.Take()
	c.Assert(queue.Len(), Equals, 0)
	c.Assert(switches, DeepEquals, [4]byte{1, 0, 1, 0})
	c.Assert(queue.Len(), Equals, 0)
}
