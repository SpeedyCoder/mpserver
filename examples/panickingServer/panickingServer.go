package main
import "mpserver"


func panickingComponent(in <-chan mpserver.Job, 
                        out chan<- mpserver.Job) {
    i := 0
    for job := range in {
        job.SetResult("Hello World!")
        if i >= 2 {
            panic("There were more than 2 requests.")
        }
        out <- job
        i++
    }
    close(out)
}

func main() {
    toComponent := mpserver.GetChan()
    toRouter := mpserver.GetChan()
    toStringWriter := mpserver.GetChan()
    toErrorWriter := mpserver.GetChan()

    phComp := mpserver.PannicHandler(panickingComponent)
    go phComp(toComponent, toRouter)
    go mpserver.ErrorRouter(
        toRouter, toStringWriter, toErrorWriter)
    go mpserver.StringWriter(toStringWriter)
    go mpserver.ErrorWriter(toErrorWriter)

    mpserver.Listen("/", toComponent, nil)
    mpserver.ListenAndServe(":3000", nil)
}