package ocrsdk

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Documentation: http://ocrsdk.com/documentation/apireference/processImage/

const (
	processTextFieldURL = "/processTextField"
)

type ProcessTextField struct {
	baseURL            string
	Region             string
	Language           string
	LetterSet          string
	RegExp             string
	TextType           string
	OneTextLine        bool
	OneWordPerTextLine bool
	MarkingType        string
	PlaceholdersCount  string
	WritingStyle       string
	Description        string
	PDFPassword        string
}

func NewProcessTextField(appId, secret string) *ProcessTextField {
	p := ProcessTextField{
		Language:           "English",
		OneTextLine:        false,
		OneWordPerTextLine: false,
		baseURL:            fmt.Sprintf(baseURL, appId, secret),
	}

	return &p
}

func (p *ProcessTextField) createURL() string {
	v := url.Values{}

	v.Set("language", p.Language)
	if p.Region != "" {
		v.Add("region", p.Region)
	}
	if p.LetterSet != "" {
		v.Add("letterSet", p.LetterSet)
	}
	if p.RegExp != "" {
		v.Add("regExp", p.RegExp)
	}
	if p.TextType != "" {
		v.Add("textType", p.TextType)
	}
	if p.OneTextLine == true {
		v.Add("correctOrientation", "true")
	}
	if p.OneWordPerTextLine == true {
		v.Add("correctSkew", "true")
	}
	if p.MarkingType != "" {
		v.Add("markingType", p.MarkingType)
	}
	if p.PlaceholdersCount != "" {
		v.Add("placeholdersCount", p.PlaceholdersCount)
	}
	if p.WritingStyle != "" {
		v.Add("writingStyle", p.WritingStyle)
	}
	if p.Description != "" {
		v.Add("description", p.Description)
	}
	if p.PDFPassword != "" {
		v.Add("pdfPassword", p.PDFPassword)
	}

	return fmt.Sprintf("%s%s?%s", p.baseURL, processTextFieldURL, v.Encode())

}

func (p *ProcessTextField) Do(pathFile string) (string, error) {
	postURL := p.createURL()
	fmt.Println(postURL)

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
