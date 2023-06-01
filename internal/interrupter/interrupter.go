package interrupter

import (
	"os"
	"os/signal"
)

type Interrupters struct {
	fs []func() error
}

type Logger interface {
	Error(args ...interface{})
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

func (i *Interrupters) Run(l Logger) {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		<-sigc

		ers := i.Do()

		if len(ers) > 0 {
			for _, v := range ers {
				l.Error(v)
			}
			os.Exit(1)
		}

		os.Exit(0)

	}()

}
