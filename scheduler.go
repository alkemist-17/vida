package vida

type Scheduler struct {
	main    *Thread
	current *Thread
	Pool    *threadPool
}

func (s *Scheduler) Current() *Thread {
	return s.current
}

func (s *Scheduler) SetCurrent(th *Thread) {
	s.current = th
}

func (s *Scheduler) Main() *Thread {
	return s.main
}

func (s *Scheduler) SetMain(th *Thread) {
	s.main = th
	s.current = th
}

func (s *Scheduler) Acquire(fn *Function, script *Script) *Thread {
	th := s.Pool.getThread()
	th.Script.MainFunction = fn
	th.State = Ready
	return th
}

func (s *Scheduler) Release(th *Thread) {
	th.State = Completed
	th.Invoker = nil
	th.Channel = GlobalNil
	s.Pool.releaseThread()
}
