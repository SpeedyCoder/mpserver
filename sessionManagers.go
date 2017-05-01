package mpserver

import(
    "log"
    "errors"
    "time"
    "crypto/rand"
    "encoding/base64"
)

// Code for id generation by Matt Silverlock

// GenerateRandomBytes returns securely generated random bytes. 
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    // Note that err == nil only if we read len(b) bytes.
    if err != nil {
        return nil, err
    }

    return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
    b, err := GenerateRandomBytes(s)
    return base64.URLEncoding.EncodeToString(b), err
}

type State interface {
    Next(val Value) (State, error)
    Terminal() bool
    Result() Any
}

func startNewSession(val Value, initial State, seshExp time.Duration, store Storage, out chan<- Value) {
    id, err := GenerateRandomString(32)
    if (err != nil) {
        // if the random generator fails
        log.Println(err)
        val.SetResult(errors.New("Session-Id generation failed"))
        out <- val
        return
    }
    log.Println("Id generated: " + id)

    state, err := initial.Next(val)
    if (err != nil) {
        val.SetResult(err)
        out <- val
        return
    }

    store.Set(id, StoreValue{state, time.Now().Add(seshExp)})
    val.SetResult(state.Result())
    val.SetHeader("Session-Id", id)
    out <- val
}

func SessionManagementComponent(store Storage, initial State, seshExp time.Duration, useCleaner bool) Component {
    return func (in <-chan Value, out chan<- Value) {
        var cleanerShutDown chan bool; 
        if (useCleaner) {
            cleanerShutDown = make(chan bool, 1)
            go StoreCleaner(store, cleanerShutDown, seshExp)  
        }

        for val := range in {
            // log.Println(val.Request.Header)
            id := val.GetRequest().Header.Get("Session-Id")
            log.Println(id)
            if (id == ""){
                log.Println("No Session-Id")
                startNewSession(val, initial, seshExp, store, out)
                continue
            }
            elem, in := store.Get(id)
            if (!in){
                log.Println("Session-Id not in store")
                // Invalid or expiredID
                startNewSession(val, initial, seshExp, store, out)
                continue
            }

            storeValue, _ := elem.(StoreValue)
            now := time.Now()
            if (storeValue.Time.After(now)) {
                state, _ := storeValue.Value.(State)
                next, err := state.Next(val)
                if (err != nil) {
                    log.Println("Error while generating next state")
                    val.SetResult(err)
                    out <- val
                    continue
                }
                if (next.Terminal()) {
                    log.Println("Terminal state")
                    // If next is a terminal state, then 
                    // the session terminates
                    store.Remove(id)
                } else {
                    log.Println("Updated state")
                    storeValue.Value = next
                    store.Set(id, storeValue)
                    val.SetHeader("Session-Id", id)
                }

                val.SetResult(next.Result())
                out <- val
            } else {
                log.Println("Session expired")
                // Session expired
                store.Remove(id)
                startNewSession(val, initial, seshExp, store, out)
            }
        }
        if (useCleaner) {
            cleanerShutDown <- true
        }
        close(out)
    }
}


