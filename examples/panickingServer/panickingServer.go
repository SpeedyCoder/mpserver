package main
import "mpserver"


func panickingComponent(in <-chan mpserver.Value, 
                        out chan<- mpserver.Value) {
    i := 0
    for val := range in {
        val.SetResult("Hello World!")
        if i >= 2 {
            panic("There were more than 2 requests.")
        }
        out <- val
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