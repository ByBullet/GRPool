package grpool

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type Task[T any] struct {
	args []T
	fn   func(args ...T)
}

func (t *Task[T]) Do() {
	t.fn(t.args...)
}

type Pool[T any] struct {
	c         chan *Task[T]
	grProfile map[string]context.CancelFunc
	wait      sync.WaitGroup
}

func newPool[T any](size int) *Pool[T] {
	p := new(Pool[T])
	p.c = make(chan *Task[T], size)
	p.grProfile = make(map[string]context.CancelFunc)

	for i := 0; i < size; i++ {
		routineName := uuid.New().String()
		ctx, cancel := context.WithCancel(context.Background())
		p.wait.Add(1)
		go func(ctx context.Context, rName string) {
			for task := range p.c {
				select {
				case <-ctx.Done():
					break
				default:
					task.Do()
				}
			}
			p.wait.Done()
			log.Println("routineName:", rName, "routine Done")
		}(ctx, routineName)
		p.grProfile[routineName] = cancel
	}
	return p
}

// NewFixedPool 固定大小的goroutine pool
func NewFixedPool[T any](size int) *Pool[T] {
	return newPool[T](size)
}

func (p *Pool[T]) AddTask(task *Task[T]) {
	p.c <- task
}

func (p *Pool[T]) Done() {
	close(p.c)
	for k, v := range p.grProfile {
		if len(p.c) < len(p.grProfile) {
			time.Sleep(time.Second / 10)
			continue
		}
		v()
		delete(p.grProfile, k)
		log.Println("routineName:", k, "routine Done Apply")
	}
	fmt.Println(len(p.grProfile))
	p.wait.Wait()
	log.Println("pool done")
}
