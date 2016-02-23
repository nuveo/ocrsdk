package ocrsdk

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	baseURL        = "http://%s:%s@cloud.ocrsdk.com"
	getTaskStatus  = "/getTaskStatus?taskid=%s"
	submitImageURL = "/submitImage"
)

// Response head of XML response
type Response struct {
	Task Task `xml:"task"`
}

// Task body of XML Response
type Task struct {
	TaskID      string `xml:"id,attr"`
	Status      string `xml:"status,attr"`
	DownloadURL string `xml:"resultUrl,attr"`
}

// Creates a new file upload http request
func newfileUploadRequest(uri, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, err
}

// ProcessUnmarshal get http response and process to struct `Response`
func ProcessUnmarshal(resp *http.Response) (Response, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return Response{}, err
	}
	defer resp.Body.Close()

	var r Response
	err = xml.Unmarshal(body, &r)
	if err != nil {
		log.Println(err)
		return Response{}, err
	}

	return r, err
}
