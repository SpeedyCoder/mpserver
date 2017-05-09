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

// Definition of methods of the State interface
func (s ShoppingCart) Next(
    val mpserver.Value) (mpserver.State, error) {
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
var rmvActionMaker = mpserver.MakeComponent(
                                        rmvActionMakerFunc)
var buyActionMaker = mpserver.MakeComponent(buyActionMakerFunc)

const SessionExpiration = time.Minute*5
const AddTimeout = time.Second
const RemoveTimeout = time.Minute*5
var InitialState = ShoppingCart{nil, false}

func actionWriter(actionMaker mpserver.Component, 
                  storage mpserver.Storage) mpserver.Writer {
    return func (in <-chan mpserver.Value) {
        // Define session component
        seshComp := mpserver.SessionManagementComponent(
                storage, InitialState, SessionExpiration)
        // Wrap it in an Error Passer
        seshComp = mpserver.ErrorPasser(seshComp)

        // Create channels
        toSeshManager := mpserver.GetChan()
        toSplitter := mpserver.GetChan()
        toWriter := mpserver.GetChan()
        toErrWriter := mpserver.GetChan()

        // Start the components and writers
        go actionMaker(in, toSeshManager)
        go seshComp(toSeshManager, toSplitter)
        go mpserver.ErrorSplitter(
            toSplitter, toWriter, toErrWriter)
        go mpserver.JsonWriter(toWriter)
        mpserver.ErrorWriter(toErrWriter)
    }
}

func main() {
    // Create the storage
    storage := mpserver.NewMemStorage()
    // Create writers
    addActionWriter := actionWriter(addActionMaker, storage)
    rmvActionWriter := actionWriter(rmvActionMaker, storage)
    buyActionWriter := actionWriter(buyActionMaker, storage)

    // Wrap them in Load Balancers
    addActionWriter = mpserver.DynamicLoadBalancerWriter(
        addActionWriter, 100, AddTimeout, RemoveTimeout)
    rmvActionWriter = mpserver.DynamicLoadBalancerWriter(
        rmvActionWriter, 20, AddTimeout, RemoveTimeout)
    buyActionWriter = mpserver.DynamicLoadBalancerWriter(
        buyActionWriter, 40, AddTimeout, RemoveTimeout)

    // Make channels
    toAddActionMaker := make(mpserver.ValueChan)
    toRmvActionMaker := make(mpserver.ValueChan)
    toBuyActionMaker := mpserver.GetChan()

    // Start the load balanced writers
    go addActionWriter(toAddActionMaker)
    go rmvActionWriter(toRmvActionMaker)
    go buyActionWriter(toBuyActionMaker)
    go mpserver.StorageCleaner(storage, nil, SessionExpiration)

    // Start the server
    mux := http.NewServeMux()
    mpserver.Listen(mux, "/add", toAddActionMaker)
    mpserver.Listen(mux, "/remove", toRmvActionMaker)
    mpserver.Listen(mux, "/buy", toBuyActionMaker)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}

