package main

import(
    "log"
    "net/http"
    "mpserver"
    "time"
)

const SessionExpiration = time.Minute*5
const AddTimeout = time.Second
const RemoveTimeout = time.Minute*5
var InitialState = ShoppingCart{nil, false}

func actionWriter(actionMaker mpserver.Component, 
                  storage mpserver.Storage) mpserver.Writer {
    return func (in <-chan mpserver.Value) {
        // Define session component
        seshComp := mpserver.SessionManager(
                storage, InitialState, SessionExpiration)
        // Wrap it in an Error Passer
        seshComp = mpserver.ErrorPasser(seshComp)

        // Create channels
        toSeshManager := mpserver.GetChan()
        toRouter := mpserver.GetChan()
        toWriter := mpserver.GetChan()
        toErrWriter := mpserver.GetChan()

        // Start the components and writers
        go actionMaker(in, toSeshManager)
        go seshComp(toSeshManager, toRouter)
        go mpserver.ErrorRouter(
            toRouter, toWriter, toErrWriter)
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
    toAddActionMaker := mpserver.GetChan()
    toRmvActionMaker := mpserver.GetChan()
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