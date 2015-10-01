package ocrsdk

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	baseURL         = "http://%s:%s@cloud.ocrsdk.com"
	processImageURL = "/processImage?language=%s&exportFormat=txt"
	getTaskStatus   = "/getTaskStatus?taskid=%s"
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

// Config struct with ApplicationID and Password
type Config struct {
	ApplicationID string
	Password      string
}

// Setup add ApplicationID and Password in Config struct
func (c *Config) Setup() {
	usr := os.Getenv("APPLICATION_ID")
	psw := os.Getenv("PASSWORD")

	if usr != "" && psw != "" {
		c.ApplicationID = usr
		c.Password = psw
	} else {
		log.Fatal("Export APPLICATION_ID and PASSWORD environ vars")
	}
}

// MakeBaseURL return url with application id and password
func (c *Config) MakeBaseURL() string {
	URI := fmt.Sprintf(baseURL, c.ApplicationID, c.Password)
	return URI
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

// Ocrsdk bridge to webapp OCRSDK
func Ocrsdk(pathFile string, language string) (string, error) {
	app := Config{}
	app.Setup()
	base := app.MakeBaseURL()

	p := fmt.Sprintf(processImageURL, language)
	postURL := fmt.Sprintf("%s%s", base, p)

	request, err := newfileUploadRequest(postURL, pathFile)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Println("Making request to", pathFile)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return "", err
	}

	r, err := ProcessUnmarshal(resp)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if r.Task.Status != "Queued" {
		return "", fmt.Errorf("Task has a problem, Task status: %s", r.Task.Status)
	}

	log.Println("Processing task!")
	time.Sleep(3 * time.Second)

	g := fmt.Sprintf(getTaskStatus, r.Task.TaskID)
	getURL := fmt.Sprintf("%s%s", base, g)

	resp, err = http.Get(getURL)
	if err != nil {
		log.Println(err)
		return "", err
	}

	r, err = ProcessUnmarshal(resp)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if r.Task.Status != "Completed" {
		return "", fmt.Errorf("Task with problem!, Task status: %s", r.Task.Status)
	}

	log.Println("Task completed!")
	resp, err = http.Get(r.Task.DownloadURL)
	if err != nil {
		log.Println(err)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	resp.Body.Close()
	return string(body), nil
}
