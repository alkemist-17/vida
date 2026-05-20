package vida

type ThreadState int

const (
	Ready ThreadState = iota
	Running
	Suspended
	Waiting
	Completed
)

const (
	minFrameSize = 8
	minStackSize = 256
	maxFrameSize = 1024
	maxStackSize = 1024
)

func (state ThreadState) String() string {
	switch state {
	case Ready:
		return "ready"
	case Running:
		return "running"
	case Suspended:
		return "suspended"
	case Waiting:
		return "waiting"
	case Completed:
		return "completed"
	default:
		return "unknown"
	}
}

type Thread struct {
	Frames  []frame
	Stack   []Value
	Script  *Script
	Frame   *frame
	Invoker *Thread
	State   ThreadState
	Channel Value
	Reg     uint64
	fp      int
}

func newMainThread(script *Script, extensionlibsloader LibsLoader) (*Thread, error) {
	extensionlibsLoader, clbu = extensionlibsloader, script.Store
	th := &Thread{
		Frames: make([]frame, frameSize),
		Stack:  make([]Value, stacksize),
		Script: script,
		State:  Running,
	}
	return th, nil
}

func newThread(fn *Function, script *Script) *Thread {
	return &Thread{
		Script: &Script{
			Konstants:    script.Konstants,
			Store:        script.Store,
			ErrorInfo:    script.ErrorInfo,
			MainFunction: fn,
		},
		Frames: make([]frame, frameSize),
		Stack:  make([]Value, stacksize),
	}
}
