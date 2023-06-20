package interrupter

import (
	"os"
	"os/signal"

	"go.uber.org/zap"
)

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

func (i *Interrupters) Run(l *zap.SugaredLogger) {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		<-sigc

		ers := i.Do()

		if len(ers) > 0 {
			for _, err := range ers {
				l.Errorf("cannot do interrrupt err: %w", err)
			}
			os.Exit(1)
		}

		os.Exit(0)

	}()

}
