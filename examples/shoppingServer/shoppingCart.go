package main

import "errors"
import "mpserver"

type ShoppingCart struct {
    items map[string]int
    bought bool
}

// Definition of methods of the State interface
func (s ShoppingCart) Next(
    job mpserver.Job) (mpserver.State, error) {
    a, ok := job.GetResult().(Action)
    if (!ok) {
        return nil, errors.New("No action provided")
    }

    return a.performAction(s)
}

func (s ShoppingCart) Result() interface{} {
    return s.items
}

func (s ShoppingCart) Terminal() bool {
    return s.bought
}