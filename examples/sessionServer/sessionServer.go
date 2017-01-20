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

func (s Session) Next(http.Request) (mpserver.State, error) {
    return Session{s.step+1, s.limit}, nil
}

func (s Session) Result() mpserver.Any {
    return "Hello world " + strconv.Itoa(s.step)
}

func (s Session) Terminal() bool {
    return s.step == s.limit
}

var initial = Session{0, 5}

func main() {
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    toSplitter := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)
    sComp := mpserver.SessionManagementComponent(initial, time.Second*60)
    go sComp(in, toSplitter)
    go mpserver.ErrorSplitter(toSplitter, out, errChan)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}



