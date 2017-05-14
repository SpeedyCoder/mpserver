package mpserver

import "reflect"

// Condition is a type that represents conditions that are used
// for splitting. It is a function the takes a job and returns
// a boolean. It shouldn't change the provided job.
type Condition func (job Job) bool

// ToOutChans converts a slice of type chan Job to a slice of
// of type []chan<- Job. That is it creates a slice that 
// contains pointers to the writing end of the provided channels.
func ToOutChans(chans [](chan Job)) []chan<- Job {
    res := make([]chan<- Job, len(chans))
    for i, ch := range chans {
        res[i] = ch
    }
    return res
}

// ToInChans converts a slice of type chan Job to a slice of
// of type []<-chan Job. That is it creates a slice that 
// contains pointers to the reading end of the provided channels.
func ToInChans(chans [](chan Job)) []<-chan Job {
    res := make([]<-chan Job, len(chans))
    for i, ch := range chans {
        res[i] = ch
    }
    return res
}

// Router takes an input channel, default output channel, a 
// slice of output channels  and a slice of conditions. The 
// number of output channels and the number of conditions should 
// be the same. For every input job the Router evaluates the 
// conditions from first to last. When a condition returns true
// the job is written to a corresponding output channel and the 
// processing of the current job terminates. If all conditions 
// return false then the job is written to the default output 
// channel.
func Router(in <-chan Job, defOut chan<- Job, 
            outs []chan<- Job, conds []Condition) {
    if len(outs) != len(conds) {
        panic("Number of channels and conditions is not equal.")
    }

    for job := range in {
        sent := false
        for i, cond := range conds {
            if cond(job) {
                outs[i] <- job; sent = true
                break
            }
        }

        if !sent {
            defOut <- job
        }
    }

    for _, ch := range outs {
        close(ch)
    }
}

// ErrorRouter reads jobs from its input channel and sends
// all jobs that have an error in the result field to the 
// error channel. It sends all other jobs to the default output
// channel.
func ErrorRouter(in <-chan Job, defOut chan<- Job, 
                   errChan chan<- Job) {
    for job := range in {
        if _, ok := job.GetResult().(error); ok {
            errChan <- job
        } else {
            defOut <- job
        }
    }
    close(defOut)
    close(errChan)
}

// Collector reads jobs from all input channels and forwards them
// to its output channel. It closes the output channel and 
// terminates only when all of the input channels have been 
// closed.
func Collector(ins []<-chan Job, out chan<- Job) {
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
            job := value.Interface().(Job)
            out <- job
        } else {
            cases = append(cases[:index], cases[index+1:]...)
            if len(cases) == 0 {
                done = true
            }
        }
    }
    close(out)
}