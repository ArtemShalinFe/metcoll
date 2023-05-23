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
		go st.runIntervalStateSaving()
	} else {
		st.logger.Info("sync state saving disabling")
	}

	st.runGracefullInterrupt()

	return st, nil

}

func (st *State) SyncSave() error {

	if st.storeInterval != 0 {
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

	st.logger.Info("try save state")

	data, err := st.stg.GetState()
	if err != nil {
		st.logger.Error("cannot get storage state err: ", err)
		return err
	}

	data = append(data, '\n')

	if _, err = file.Write(data); err != nil {
		return nil
	}

	st.logger.Info("state was saved")

	return nil

}

func (st *State) Load() error {

	file, err := os.OpenFile(st.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		st.logger.Error("cannot open or creating file for state loading err: ", err)
		return err
	}

	st.logger.Info("try restoring state")

	defer file.Close()

	r := bufio.NewReader(file)
	b, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			st.logger.Error("cannot reading file for state loading err: ", err)
			return err
		}
	}

	if len(b) == 0 {
		st.logger.Info("state cannot be restored, file is empty")
		return nil
	}

	if err = st.stg.SetState(b); err != nil {
		st.logger.Error("cannot set state err: ", err)
		return err
	}

	st.logger.Info("state was restored")

	return nil

}

func (st *State) runIntervalStateSaving() {

	sleepDuration := time.Duration(st.storeInterval) * time.Second

	for {
		time.Sleep(sleepDuration)
		if err := st.Save(); err != nil {
			st.logger.Error("cannot save state err: ", err)
		}
	}

}

func (st *State) runGracefullInterrupt() {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		sig := <-sigc
		st.logger.Info("incomming signal ", sig)

		if err := st.Save(); err != nil {
			st.logger.Error("cannot save state err: ", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}

	}()

}
