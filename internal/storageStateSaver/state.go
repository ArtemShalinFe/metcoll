package storageStateSaver

import (
	"bufio"
	"os"
)

type State struct {
	fileStoragePath string
}

func NewState(fileStoragePath string) *State {

	return &State{
		fileStoragePath: fileStoragePath,
	}

}

func (st *State) Save(data []byte) error {

	file, err := os.OpenFile(st.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data = append(data, '\n')

	_, err = file.Write(data)
	return err

}

func (st *State) Load() ([]byte, error) {

	file, err := os.OpenFile(st.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	r := bufio.NewReader(file)
	return r.ReadBytes('\n')

}
