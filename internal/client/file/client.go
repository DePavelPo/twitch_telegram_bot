package file

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type FileClient struct {
}

func NewFileClient() *FileClient {
	return &FileClient{}
}

func (fc *FileClient) DownloadFile(ctx context.Context, URL, fileName string) error {

	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.Errorf("client failed with status code %d", response.StatusCode)
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil

}
