package distributor

import (
	"bytes"
	"os"
	"path"
)

type LocalDistributor struct {
	Directory string
	FileName  string
}

func (d LocalDistributor) Distribute(content *bytes.Buffer) error {
	if _, err := os.Stat(d.Directory); os.IsNotExist(err) {
		if err := os.Mkdir(d.Directory, os.ModePerm); err != nil {
			return err
		}
	}

	filePath := path.Join(d.Directory, d.FileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	content.WriteTo(file)

	return nil
}
