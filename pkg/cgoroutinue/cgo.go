package cgoroutinue

import "github.com/panjf2000/ants/v2"

var GoroutinePool *ants.Pool

func InitGoroutinePool(size int) {
	GoroutinePool = NewGoroutinePool(size)
}

func NewGoroutinePool(size int) *ants.Pool {
	p, _ := ants.NewPool(size)
	return p
}
