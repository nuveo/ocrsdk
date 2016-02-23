package ocrsdk

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Documentation: http://ocrsdk.com/documentation/apireference/processFields/

const (
	processFieldsURL = "/processFields"
)

type ProcessFields struct {
	baseURL                  string
	TaskID                   string
	Description              string
	WriteRecognitionVariants string
}

func NewProcessFields(appId, secret string) *ProcessFields {
	p := ProcessFields{
		baseURL: fmt.Sprintf(baseURL, appId, secret),
	}

	return &p
}

func (p *ProcessFields) createURL() string {
	v := url.Values{}

	v.Set("taskid", p.TaskID)
	if p.Description != "" {
		v.Add("description", p.Description)
	}
	if p.WriteRecognitionVariants != "" {
		v.Add("writeRecognitionVariants", p.WriteRecognitionVariants)
	}

	return fmt.Sprintf("%s%s?%s", p.baseURL, processFieldsURL, v.Encode())

}

func (p *ProcessFields) Do(xmlFile, pathFile string) (string, error) {
	//First submit image calling submitImage endpoint

	request, err := newfileUploadRequest(p.baseURL+submitImageURL, pathFile)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Println("Making request to ", pathFile)

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

	log.Println("Processing task!")
	time.Sleep(3 * time.Second)

	g := fmt.Sprintf(getTaskStatus, r.Task.TaskID)
	getURL := fmt.Sprintf("%s%s", p.baseURL, g)

	for {
		log.Println("Getting Task status")
		var stop bool
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

		switch r.Task.Status {
		case "InProgress":
			log.Println("Task In Progress")
			time.Sleep(5 * time.Second)
		case "Submitted":
			log.Println("Task Completed!")
			stop = true
		case "ProcessingFailed", "NotEnoughCredits":
			log.Println("Task Failed!")
			return "", fmt.Errorf("Task status: %s", r.Task.Status)
		default:
			log.Println("waiting...")
			time.Sleep(5 * time.Second)
		}

		if stop == true {
			break
		}
	}

	//Get taskID to send xml request
	p.TaskID = r.Task.TaskID

	//Add xmlfile in body and send request

	postURL := p.createURL()

	requestXML, err := newfileUploadRequest(postURL, xmlFile)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Println("Making request to ", xmlFile)

	client = &http.Client{}
	resp, err = client.Do(requestXML)
	if err != nil {
		log.Println(err)
		return "", err
	}

	r, err = ProcessUnmarshal(resp)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if r.Task.Status != "Queued" {
		return "", fmt.Errorf("Task has a problem, Task status: %s", r.Task.Status)
	}

	log.Println("Processing task!")
	time.Sleep(3 * time.Second)

	g = fmt.Sprintf(getTaskStatus, r.Task.TaskID)
	getURL = fmt.Sprintf("%s%s", p.baseURL, g)

	for {
		log.Println("Getting Task status")
		var stop bool
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

		switch r.Task.Status {
		case "InProgress":
			log.Println("Task In Progress")
			time.Sleep(5 * time.Second)
		case "Completed":
			log.Println("Task Completed!")
			stop = true
		case "ProcessingFailed", "NotEnoughCredits":
			log.Println("Task Failed!")
			return "", fmt.Errorf("Task status: %s", r.Task.Status)
		default:
			log.Println("waiting...")
			time.Sleep(5 * time.Second)
		}

		if stop == true {
			break
		}
	}

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
