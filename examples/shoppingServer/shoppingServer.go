package main

import(
    "log"
    "errors"
    "net/http"
    "mpserver"
    "time"
    "strings"
)

// Type definitions
type ShoppingCart struct {
    items map[string]int
    bought bool
}

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

// Definitions of performAction method
func (a AddAction) performAction(s ShoppingCart) 
                                (mpserver.State, error) {
    if (s.items == nil) {
        s.items = make(map[string]int)
    }

    for _, item := range a.items {
        s.items[item] += 1
    }

    return s, nil
}

func (a RemoveAction) performAction(s ShoppingCart) 
                                (mpserver.State, error) {
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

func (a BuyAction) performAction(s ShoppingCart) 
                                (mpserver.State, error) {
    // In real application this is where the payment and writing
    // to the database would happen
    s.bought = true
    return s, nil
}

// Definition of methods of the State interface
func (s ShoppingCart) Next(val mpserver.Value) 
                                (mpserver.State, error) {
    a, ok := val.GetResult().(Action)
    if (!ok) {
        return nil, errors.New("Action not provided")
    }

    return a.performAction(s)
}

func (s ShoppingCart) Result() interface{} {
    return s.items
}

func (s ShoppingCart) Terminal() bool {
    return s.bought
}

// Definition of Action makers
func addActionMakerFunc(val mpserver.Value) mpserver.Value {
    q := val.GetRequest().URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        items = strings.Split(items[0], ",")
    }
    val.SetResult(AddAction{items})
    
    return val
}

func removeActionMakerFunc(val mpserver.Value) mpserver.Value {
    q := val.GetRequest().URL.Query()
    items, ok := q["item"]
    if (ok && len(items) > 0) {
        val.SetResult(RemoveAction{items[0]})
    } else {
        val.SetResult(
            errors.New("removeActionMaker: Item not provided."))
    }
    
    return val
}

func buyActionMakerFunc(val mpserver.Value) mpserver.Value {
    val.SetResult(BuyAction{})
    return val
}

var addActionMaker = mpserver.MakeComponent(addActionMakerFunc)
var removeActionMaker = mpserver.MakeComponent(
                                        removeActionMakerFunc)
var buyActionMaker = mpserver.MakeComponent(buyActionMakerFunc)

func main() {
    // Make channels
    toAddActionMaker := make(mpserver.ValueChan)
    toRemoveActionMaker := make(mpserver.ValueChan)
    toBuyActionMaker := mpserver.GetChan()
    in := mpserver.GetChan()
    out := mpserver.GetChan()
    toWriter := mpserver.GetChan()
    errChan := mpserver.GetChan()

    // Start action makers
    go addActionMaker(toAddActionMaker, in)
    go removeActionMaker(toRemoveActionMaker, in)
    go buyActionMaker(toBuyActionMaker, in)

    // Start the session manager
    initial := ShoppingCart{nil, false}
    store := mpserver.NewMemStore()
    sComp := mpserver.SessionManagementComponent(
                store, initial, time.Second*60*5, true)
    sComp = mpserver.ErrorPasser(sComp)
    go sComp(in, out)

    // Start Error Splitter and Writers
    go mpserver.ErrorSplitter(out, toWriter, errChan)
    go mpserver.JsonWriter(toWriter)
    go mpserver.ErrorWriter(errChan)

    // Start the server
    mux := http.NewServeMux()
    mpserver.Listen(mux, "/add", toAddActionMaker)
    mpserver.Listen(mux, "/remove", toRemoveActionMaker)
    mpserver.Listen(mux, "/buy", toBuyActionMaker)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}

