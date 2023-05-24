package interrupter

type Interrupters struct {
	fs []func() error
}

func NewInterrupters() *Interrupters {
	return &Interrupters{}
}

func (i *Interrupters) Use(f func() error) {
	i.fs = append(i.fs, f)
}

func (i *Interrupters) Do() []error {

	var ers []error

	for _, f := range i.fs {
		if err := f(); err != nil {
			ers = append(ers, err)
		}
	}

	return ers

}
