package utils

import (
	"math/rand"
	"time"
)

type Random interface {
	Rand(min, max int) int
}

type RandomImpl struct{}

func NewRandomImpl() *RandomImpl {
	return &RandomImpl{}
}

func (r *RandomImpl) Rand(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

type MockRandomImpl struct {
	val int
}

func NewMockRandomImpl() *MockRandomImpl {
	return &MockRandomImpl{}
}

func (r *MockRandomImpl) Rand(min, max int) int {
	return r.val
}

func (r *MockRandomImpl) SetVal(val int) {
	r.val = val
}
