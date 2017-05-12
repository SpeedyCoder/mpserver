package main

import(
    "log"
    "strconv"
    "net/http"
    "mpserver"
    "time"
)

type Session struct {
    step int
    limit int
}

func (s Session) Next(val mpserver.Value) (mpserver.State, error) {
    return Session{s.step+1, s.limit}, nil
}

func (s Session) Result() interface{} {
    return "Hello world " + strconv.Itoa(s.step)
}

func (s Session) Terminal() bool {
    return s.step == s.limit
}

var initial = Session{0, 5}

func main() {
    in := mpserver.GetChan()
    out := mpserver.GetChan()
    toStringWriter := mpserver.GetChan()
    toErrorWriter := mpserver.GetChan()

    store := mpserver.NewMemStorage()
    sComp := mpserver.SessionManagementComponent(store, initial, time.Second*15)
    go sComp(in, out)
    go mpserver.ErrorSplitter(in, toStringWriter, toErrorWriter)
    go mpserver.StringWriter(toStringWriter)
    go mpserver.ErrorWriter(toErrorWriter)

    mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}



