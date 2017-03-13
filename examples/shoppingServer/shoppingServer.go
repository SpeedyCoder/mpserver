package main

import(
    "fmt"
    "log"
    "errors"
    "net/http"
    "mpserver"
    "time"
    "strings"
)

type Session struct {
    items map[string]int
    finished bool
}

type Action interface {
    performAction(s Session) (mpserver.State, error)
}

type AddAction struct {
    items []string
}

type RemoveAction struct {
    item string
}

type BuyAction struct {}

func (a AddAction) performAction(s Session) (mpserver.State, error) {
    if (s.items == nil) {
        s.items = make(map[string]int)
    }

    for _, item := range a.items {
        s.items[item] += 1
    }

    return s, nil
}

func (a RemoveAction) performAction(s Session) (mpserver.State, error) {
    if (s.items == nil) {
        return nil, errors.New("Shopping cart is empty.")
    }

    _, ok := s.items[a.item]
    if (!ok) {
        return nil, errors.New("Item: "+a.item+" is not in the cart")
    }
    s.items[a.item] -= 1

    if (s.items[a.item] == 0) {
        delete(s.items, a.item)
    }

    return s, nil
}

func (a BuyAction) performAction(s Session) (mpserver.State, error) {
    // In real application this is where the payment and writing
    // to the database would happen
    s.finished = true
    return s, nil
}



func (s Session) Next(val mpserver.Value) (mpserver.State, error) {
    a, ok := val.Result.(Action)
    if (!ok) {
        return nil, errors.New("Action not provided")
    }

    return a.performAction(s)
}

func (s Session) Result() mpserver.Any {
    if (s.finished) {
        return fmt.Sprintf("Bought: %v", s.items)
    }
    return s.items
}

func (s Session) Terminal() bool {
    return s.finished
}

func addActionMaker(val mpserver.Value) mpserver.Value {
    q := val.Request.URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        items = strings.Split(items[0], ",")
    }
    val.Result = AddAction{items}
    
    return val
}

func removeActionMaker(val mpserver.Value) mpserver.Value {
    q := val.Request.URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        val.Result = RemoveAction{items[0]}
    } else {
        val.Result = errors.New("removeActionMaker: Item not provided.")
    }
    
    return val
}

func buyActionMaker(val mpserver.Value) mpserver.Value {
    val.Result = BuyAction{}
    return val
}


var initial = Session{nil, false}

func main() {
    toAddActionMaker := make(mpserver.ValueChan)
    toRemoveActionMaker := make(mpserver.ValueChan)
    toBuyActionMaker := mpserver.GetChan()

    in := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)

    go mpserver.MakeComponent(addActionMaker)(toAddActionMaker, in)
    go mpserver.MakeComponent(removeActionMaker)(toRemoveActionMaker, in)
    go mpserver.MakeComponent(buyActionMaker)(toBuyActionMaker, in)

    sComp := mpserver.SessionManagementComponent(initial, time.Second*60*5)
    sComp = mpserver.ErrorPasser(sComp)
    go sComp(in, out)
    go mpserver.AddErrorSplitter(mpserver.JsonWriter)(out, errChan)
    go mpserver.ErrorWriter(errChan)

    mux := http.NewServeMux()
    mpserver.Listen(mux, "/add", toAddActionMaker)
    mpserver.Listen(mux, "/remove", toRemoveActionMaker)
    mpserver.Listen(mux, "/buy", toBuyActionMaker)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}



