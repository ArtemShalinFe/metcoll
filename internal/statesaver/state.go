package statesaver

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/signal"
	"time"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type State struct {
	fileStoragePath string
	storeInterval   int
	stg             StorageState
	logger          Logger
}

type StorageState interface {
	GetState() ([]byte, error)
	SetState([]byte) error
}

func NewState(stg StorageState, l Logger, fileStoragePath string, storeInterval int, restore bool) (*State, error) {

	st := &State{
		fileStoragePath: fileStoragePath,
		storeInterval:   storeInterval,
		stg:             stg,
		logger:          l,
	}

	if restore {
		if err := st.Load(); err != nil {
			return nil, err
		}
	}

	if st.storeInterval != 0 {
		st.runIntervalStateSaving()
	}

	st.runGracefullInterrupt()

	return st, nil

}

func (st *State) SyncSave() error {

	if st.storeInterval != 0 {
		st.logger.Info("sync state saving disabling")
		return nil
	}

	return st.Save()

}

func (st *State) Save() error {

	file, err := os.OpenFile(st.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		st.logger.Error("cannot open or creating file for state saving err: ", err)
		return err
	}
	defer file.Close()

	data, err := st.stg.GetState()
	if err != nil {
		st.logger.Error("cannot get storage state err: ", err)
		return err
	}

	data = append(data, '\n')

	_, err = file.Write(data)
	return err

}

func (st *State) Load() error {

	file, err := os.OpenFile(st.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		st.logger.Error("cannot open or creating file for state loading err: ", err)
		return err
	}

	defer file.Close()

	r := bufio.NewReader(file)
	b, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			st.logger.Error("cannot reading file for state loading err: ", err)
			return err
		}
	}

	return st.stg.SetState(b)

}

func (st *State) runIntervalStateSaving() {

	sleepDuration := time.Duration(st.storeInterval) * time.Second
	go func() {
		for {
			if err := st.Save(); err != nil {
				st.logger.Error("cannot save state err: ", err)
			}
			time.Sleep(sleepDuration)
		}
	}()

}

func (st *State) runGracefullInterrupt() {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		sig := <-sigc
		st.logger.Error("incomming signal ", sig)

		if err := st.Save(); err != nil {
			st.logger.Error("cannot save state err: ", err)
			os.Exit(1)
		} else {
			st.logger.Info("state was saved")
			os.Exit(0)
		}
	}()

}
