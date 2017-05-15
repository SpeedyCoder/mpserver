package mpserver

import(
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

// State is a type that represents the current state of a session
// for a single user.
type State interface {
    // Next returns the next State of the session for the given 
    // user or an error if next State cannot be generated.
    Next(job Job) (State, error)

    // Terminal returns a boolean that indicates whether this 
    // State is a terminal state.
    Terminal() bool

    // Result returns a result that should be returned to the 
    // user. This should only be called after the state was 
    // updated.
    Result() interface{}
}

// startNewSession is a helper function that starts a new session
// for the given job. It generates a new session id and the 
// current state of the session from the initial state. The 
// mapping from the generated state is stored in the storage 
// object and the result of the call to the Result function on
// the current state is stored in the result field of the job,
// before it is sent to the out channel.
func startNewSession(job Job, initial State, 
    seshExp time.Duration, storage Storage, out chan<- Job) {
    id, err := GenerateRandomString(32)
    if (err != nil) {
        // The random generator failed.
        job.SetResult(errors.New("Session-Id generation failed"))
        out <- job
        return
    }

    // Generate the current state.
    state, err := initial.Next(job)
    if (err != nil) {
        job.SetResult(err)
        out <- job
        return
    }

    // Store the mapping from the id to the current state.
    storage.Set(id, StorageValue{state, time.Now().Add(seshExp)})
    job.SetResult(state.Result())
    job.SetHeader("Session-Id", id)
    out <- job
}

// SessionManager returns a component that performs session
// management.
func SessionManager(storage Storage, initial State, 
                    seshExp time.Duration) Component {
    noExpiration := seshExp <= 0
    return func (in <-chan Job, out chan<- Job) {
        for job := range in {
            id := job.GetRequest().Header.Get("Session-Id")
            if (id == ""){
                // No Session-Id was provided.
                startNewSession(
                    job, initial, seshExp, storage, out)
                continue
            }
            storageValue, in := storage.Get(id)
            if (!in){
                // Session-Id is not in the storage, hence it is 
                // either invalid or it expired and was removed 
                // from the storage.
                startNewSession(
                    job, initial, seshExp, storage, out)
                continue
            }

            now := time.Now()
            if (noExpiration || storageValue.Time.After(now)) {
                // Session hasn't expired yet or sessions don't 
                // expire.
                state, _ := storageValue.Value.(State)
                next, err := state.Next(job)
                if (err != nil) {
                    // Can't generate the next state.
                    job.SetResult(err)
                    out <- job
                    continue
                }
                if (next.Terminal()) {
                    // Next state is a terminal state, so the 
                    // session terminates.
                    storage.Remove(id)
                } else {
                    // Update the state in the storage, as 
                    // current state is not terminal.
                    storageValue.Time = time.Now().Add(seshExp)
                    storageValue.Value = next
                    storage.Set(id, storageValue)
                    job.SetHeader("Session-Id", id)
                }

                job.SetResult(next.Result())
                out <- job
            } else {
                // Session expired, so try to start a new session
                // for this user.
                storage.Remove(id)
                startNewSession(
                    job, initial, seshExp, storage, out)
            }
        }
        close(out)
    }
}