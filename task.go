package vida

import (
	"sync"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationTask() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["concepts"] = GFn(taksConcepts)
	m.Value["par"] = GFn(taskRunParallel)
	return m
}

func taksConcepts(args ...Value) (Value, error) {
	var c string
	c = `
	

	Parallelism means running a program on multiple processors, 
	with the goal of improving performance.
	Ideally, this should be done invisibly, 
	and with no semantic changes.
		
	A parallel functional program uses multiple processors to gain performance.
	For example, it may be faster to evaluate e1 + e2 by evaluating e1 and e2 in parallel, 
	and then add the results. 
	
	Parallelism has no semantic impact at all:
	the meaning of a program is unchanged whether it is executed sequentially or in parallel.
	Furthermore, the results are deterministic;
	there is no possibility that a parallel program will give one result in one run 
	and a different result in a different run.

	In contrast, concurrent program has concurrency as part of its specification.
	The program must run concurrent threads, each of which can independently perform input/output.
	The program may be run on many processors, or on one — that is an implementation choice.
	The behaviour of the program is, necessarily and by design, non-deterministic.
	Hence, unlike parallelism, concurrency has a substantial semantic impact.

	Concurrency means implementing a program by using multiple I/O-performing threads. 
	the primary goal of using concurrency is not to gain performance, 
	While a concurrent Haskell program can run on a parallel machine, 
	Since the threads perform I/O, the semantics of the program is necessarily non-deterministic.
	but rather because that is the simplest and most direct way to write the program.


	`
	return &String{Value: c}, nil
}

func taskRunParallel(args ...Value) (Value, error) {
	if len(args) > 0 {
		if A, ok := args[0].(*Array); ok && len(A.Value) > 0 {
			var wg sync.WaitGroup
			result := &Array{Value: make([]Value, len(A.Value))}
			for i := range A.Value {
				if T, ok := A.Value[i].(*Array); ok && len(T.Value) > 0 {
					switch fn := T.Value[0].(type) {
					case *Function:
						wg.Go(func() {
							th := newThread(fn, ((*clbu)[globalStateIndex].(*GlobalState)).Script, fullStack)
							vm := &VM{th}
							_, err := vm.runThread(vm.fp, 0, true, T.Value[1:]...)
							if err == nil {
								result.Value[i] = vm.Channel
							} else {
								result.Value[i] = Error{Message: &String{Value: err.Error()}}
							}
						})
					case GFn:
						wg.Go(func() {
							val, err := fn.Call(T.Value[1:]...)
							if err == nil {
								result.Value[i] = val
							} else {
								result.Value[i] = Error{Message: &String{Value: err.Error()}}
							}
						})
					default:
						wg.Wait()
						result = nil
						return NilValue, verror.ErrParallelFn
					}
				} else {
					wg.Wait()
					result = nil
					return NilValue, verror.ErrParallelArgs
				}
			}
			wg.Wait()
			return result, nil
		}
	}
	return NilValue, nil
}
