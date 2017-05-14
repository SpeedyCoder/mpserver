package mpserver 

import (
    "net/http"
    "os"
    "strings"
    "errors"
)

// Component is a generic part of the pipeline with an input and 
// an output channel. It reads jobs from its input channel, 
// processes them and writes the updated jobs to the output 
// channel.
type Component func (in <-chan Job, out chan<- Job)

// ComponentFunc represents a transition function that is applied
// to input jobs to produce output jobs. It is used with 
// MakeComponent function to generate components.
type ComponentFunc func (job Job)

// MakeComponent takes a ComponentFunc f and returns a Component
// that satisfies the contract for Components. The generated 
// component applies the function f to all input jobs before
// outputting them.
func MakeComponent(f ComponentFunc) Component {
    return func (in <-chan Job, out chan<- Job) {
        for job := range in {
            f(job); out <- job
        }
        close(out)
    }
}

// LinkComponents takes any number of components and returns a 
// component that behaves as their linear combination. That is as 
// a pipeline constructed from these components in the order in 
// which they are provided.
func LinkComponents(components ...Component) Component {
    return func (in <-chan Job, out chan<- Job) {
        iters := len(components) - 1
        if (iters == 0) {
            components[0](in, out)
        } else {
            current := GetChan()
            go components[0](in, current)
            for i:=1; i<iters; i++ {
                next := GetChan()
                go components[i](current, next)
                current = next
            }
            components[iters](current, out)
        }
        
    }
}

// ConstantComponent takes a value c of any type and returns a 
// component that writes c to the result field of all input 
// jobs and then outputs them.
func ConstantComponent(c interface{}) Component {
    return MakeComponent(func (job Job) {
        job.SetResult(c)
    })
}

// PathMaker is component generator that takes a path to a 
// directory and a prefix and returns a component. The generated
// component strips the provided prefix from the URL path of a
// request of all input jobs. Then it prepends the directory
// path to the result and writes this to the result field of
// the job.
func PathMaker(dir, prefix string) Component {
    return MakeComponent(func (job Job) {
        job.SetResult(dir + strings.TrimPrefix(
            job.GetRequest().URL.Path, prefix))
    })
}

// File Component is a Component that tries to open the file with
// a path that is stored in the result field of the input job.
// If it succeeds it stores the file handle in the result field
// of that job and outputs it. If it doesn't succeed it stores
// appropriate error in the result field.
func FileComponent (in <-chan Job, out chan<- Job) {
    handleError := func (job Job, err error) {
        job.SetResponseCode(http.StatusBadRequest)
        job.SetResult(err)
        out <- job
    }

    for job := range in {
        // Check if a path is provided
        path, ok := job.GetResult().(string)
        if (!ok) {
            job.SetResult(errors.New(
                "No path provided to FileComponent."))
            out <- job
            continue
        }

        // Check if a file with the given path exists
        if _, err := os.Stat(path); os.IsNotExist(err) {
            handleError(job, err)
            continue
        }

        // Try to open the file
        file, err := os.Open(path)
        if (err != nil) {
            handleError(job, err)
            continue
        }

        job.SetResult(file)
        out <- job        
    }
    close(out)
}