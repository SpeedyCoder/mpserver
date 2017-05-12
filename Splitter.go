package mpserver

import "reflect"

// Condition is a type that represents conditions that are used
// for splitting. It is a function the takes a Value and returns
// a boolean. It shouldn't change the provided value.
type Condition func (val Value) bool

// ToOutChans converts a slice of type chan Value to a slice of
// of type []chan<- Value. That is it creates a slice that 
// contains pointers to the writing end of the provided channels.
func ToOutChans(chans [](chan Value)) []chan<- Value {
    res := make([]chan<- Value, len(chans))
    for i, ch := range chans {
        res[i] = ch
    }
    return res
}

// ToInChans converts a slice of type chan Value to a slice of
// of type []<-chan Value. That is it creates a slice that 
// contains pointers to the reading end of the provided channels.
func ToInChans(chans [](chan Value)) []<-chan Value {
    res := make([]<-chan Value, len(chans))
    for i, ch := range chans {
        res[i] = ch
    }
    return res
}

// Splitter takes an input channel, default output channel, a 
// slice of output channels  and a slice of conditions. The 
// number of output channels and the number of conditions should 
// be the same. For every input value the Splitter evaluates the 
// conditions from first to last. When a condition returns true
// the value is written to a corresponding output channel and the 
// processing of the current value terminates. If all conditions 
// return false then the value is written to the default output 
// channel.
func Splitter(in <-chan Value, defOut chan<- Value, 
              outs []chan<- Value, conds []Condition) {
    if len(outs) != len(conds) {
        panic("Number of channels and conditions is not equal.")
    }

    for val := range in {
        sent := false
        for i, cond := range conds {
            if cond(val) {
                outs[i] <- val; sent = true
                break
            }
        }

        if !sent {
            defOut <- val
        }
    }

    for _, ch := range outs {
        close(ch)
    }
}

// ErrorSplitter reads values from its input channel and sends
// all values that have an error in the result field to the 
// error channel. It sends all other values to the default output
// channel.
func ErrorSplitter(in <-chan Value, defOut chan<- Value, 
                   errChan chan<- Value) {
    for val := range in {
        if _, ok := val.GetResult().(error); ok {
            errChan <- val
        } else {
            defOut <- val
        }
    }
    close(defOut)
    close(errChan)
}

// Collector reads input values from all input channels and 
// forwards them to its output channel. It closes the output 
// channel and terminates only when all of the input channels
// have been closed.
func Collector(ins []<-chan Value, out chan<- Value) {
    cases := make([]reflect.SelectCase, len(ins))
    for i := 0; i < len(ins); i++ {
        cases[i] = reflect.SelectCase{
            Dir: reflect.SelectRecv, 
            Chan: reflect.ValueOf(ins[i])}
    }

    done := false
    for !done {
        index, value, ok := reflect.Select(cases)
        if (ok) {
            val := value.Interface().(Value)
            out <- val
        } else {
            cases = append(cases[:index], cases[index+1:]...)
            if len(cases) == 0 {
                done = true
            }
        }
    }
    close(out)
}