package main

import "errors"
import "strings"
import "mpserver"

// Definition of Action makers
func addActionMakerFunc(val mpserver.Value) {
    q := val.GetRequest().URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        items = strings.Split(items[0], ",")
    }
    val.SetResult(AddAction{items})
}

func rmvActionMakerFunc(val mpserver.Value) {
    q := val.GetRequest().URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        val.SetResult(RemoveAction{items[0]})
    } else {
        val.SetResult(
            errors.New("removeActionMaker: Item not provided."))
    } 
}

func buyActionMakerFunc(val mpserver.Value) {
    val.SetResult(BuyAction{})
}

var addActionMaker = mpserver.MakeComponent(addActionMakerFunc)
var rmvActionMaker = mpserver.MakeComponent(rmvActionMakerFunc)
var buyActionMaker = mpserver.MakeComponent(buyActionMakerFunc)