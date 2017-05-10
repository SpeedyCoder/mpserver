package main

import "errors"
import "mpserver"

type Action interface {
    performAction(s ShoppingCart) (mpserver.State, error)
}

type AddAction struct {
    items []string
}

type RemoveAction struct {
    item string
}

type BuyAction struct {}

// Definitions of performAction methods
func (a AddAction) performAction(
    s ShoppingCart) (mpserver.State, error) {
    if (s.items == nil) {
        s.items = make(map[string]int)
    }

    for _, item := range a.items {
        s.items[item] += 1
    }

    return s, nil
}

func (a RemoveAction) performAction(
    s ShoppingCart) (mpserver.State, error) {
    if (s.items == nil) {
        return nil, errors.New("Shopping cart is empty.")
    }

    _, ok := s.items[a.item]
    if (!ok) {
        return nil, errors.New(
            "Item: "+a.item+" is not in the cart")
    }
    s.items[a.item] -= 1

    if (s.items[a.item] == 0) {
        delete(s.items, a.item)
    }

    return s, nil
}

func (a BuyAction) performAction(
    s ShoppingCart) (mpserver.State, error) {
    // In real application this is where the payment and writing
    // to the database would happen
    s.bought = true
    return s, nil
}