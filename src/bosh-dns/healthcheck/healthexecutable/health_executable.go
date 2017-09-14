package healthexecutable

type HealthExecutable struct {
}

func NewHealthExecutable() *HealthExecutable {
	return &HealthExecutable{}
}

func (he *HealthExecutable) Status() bool {
	return true
}
