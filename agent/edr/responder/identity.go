package responder

func (r *Responder) SuspendIdentity(pid int, id int64) error {
	if s, ok := r.susp.(interface{ SuspendIdentity(int, int64) error }); ok {
		return s.SuspendIdentity(pid, id)
	}
	return ErrUnsupported
}
func (r *Responder) ResumeIdentity(pid int, id int64) error {
	if s, ok := r.susp.(interface{ ResumeIdentity(int, int64) error }); ok {
		return s.ResumeIdentity(pid, id)
	}
	return ErrUnsupported
}
