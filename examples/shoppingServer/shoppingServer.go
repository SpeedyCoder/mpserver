package main

import "mpserver"
import "time"


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
    // Make channels
    toAddActionWriter := mpserver.GetChan()
    toRmvActionWriter := mpserver.GetChan()
    toBuyActionWriter := mpserver.GetChan()

    // Set up listeners
    mpserver.Listen("/add", toAddActionWriter, nil)
    mpserver.Listen("/remove", toRmvActionWriter, nil)
    mpserver.Listen("/buy", toBuyActionWriter, nil)

    // Create the storage
    storage := mpserver.NewMemStorage()

    // Create writers that share the storage object
    addActionWriter := actionWriter(addActionMaker, storage)
    rmvActionWriter := actionWriter(rmvActionMaker, storage)
    buyActionWriter := actionWriter(buyActionMaker, storage)

    // Wrap the writers in load balancers
    addActionWriter = mpserver.DynamicLoadBalancerWriter(
        addActionWriter, 100, AddTimeout, RemoveTimeout)
    rmvActionWriter = mpserver.DynamicLoadBalancerWriter(
        rmvActionWriter, 20, AddTimeout, RemoveTimeout)
    buyActionWriter = mpserver.DynamicLoadBalancerWriter(
        buyActionWriter, 40, AddTimeout, RemoveTimeout)
    
    // Start the load balanced writers and storage cleaner
    go addActionWriter(toAddActionWriter)
    go rmvActionWriter(toRmvActionWriter)
    go buyActionWriter(toBuyActionWriter)
    go mpserver.StorageCleaner(storage, nil, SessionExpiration)

    // Start the server
    mpserver.ListenAndServe(":3000", nil)
}