package main

import "errors"
import "strings"
import "mpserver"

// Definition of Action makers
func addActionMakerFunc(job mpserver.Job) {
    q := job.GetRequest().URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        items = strings.Split(items[0], ",")
    }
    job.SetResult(AddAction{items})
}

func rmvActionMakerFunc(job mpserver.Job) {
    q := job.GetRequest().URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        job.SetResult(RemoveAction{items[0]})
    } else {
        job.SetResult(
            errors.New("removeActionMaker: Item not provided."))
    } 
}

func buyActionMakerFunc(job mpserver.Job) {
    job.SetResult(BuyAction{})
}

var addActionMaker = mpserver.MakeComponent(addActionMakerFunc)
var rmvActionMaker = mpserver.MakeComponent(rmvActionMakerFunc)
var buyActionMaker = mpserver.MakeComponent(buyActionMakerFunc)