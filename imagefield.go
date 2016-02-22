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
	processImageURL = "/processImage"
)

type ProcessImage struct {
	baseURL                     string
	Language                    string
	Profile                     string
	TextType                    string
	ImageSource                 string
	CorrectOrientation          bool
	CorrectSkew                 bool
	ReadBarcodes                bool
	ExportFormat                string
	XMLWriteRecognitionVariants bool
	PDFWriteTags                string
	Description                 string
	PDFPassword                 string
}

func NewProcessImage(appId, secret string) *ProcessImage {
	pi := ProcessImage{
		Language:           "English",
		CorrectOrientation: true,
		CorrectSkew:        true,
		ReadBarcodes:       false,
		ExportFormat:       "txt",
		baseURL:            fmt.Sprintf(baseURL, appId, secret),
	}

	return &pi
}

func (pi *ProcessImage) createURL() string {
	v := url.Values{}

	v.Set("language", pi.Language)
	if pi.Profile != "" {
		v.Add("profile", pi.Profile)
	}
	if pi.TextType != "" {
		v.Add("textType", pi.TextType)
	}
	if pi.ImageSource != "" {
		v.Add("imageSource", pi.ImageSource)
	}
	if pi.CorrectOrientation == false {
		v.Add("correctOrientation", "false")
	}
	if pi.CorrectSkew == false {
		v.Add("correctSkew", "false")
	}
	if pi.ReadBarcodes == true {
		v.Add("readBarcodes", "true")
	}
	if pi.ExportFormat != "" {
		v.Add("exportFormat", pi.ExportFormat)
	}
	if pi.XMLWriteRecognitionVariants == true {
		v.Add("xml:writeRecognitionVariants", "true")
	}
	if pi.PDFWriteTags != "" {
		v.Add("pdf:writeTags", pi.PDFWriteTags)
	}
	if pi.Description != "" {
		v.Add("description", pi.Description)
	}
	if pi.PDFPassword != "" {
		v.Add("pdfPassword", pi.PDFPassword)
	}

	return fmt.Sprintf("%s%s?%s", pi.baseURL, processImageURL, v.Encode())

}

func (pi *ProcessImage) Do(pathFile string) (string, error) {
	postURL := pi.createURL()
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
	getURL := fmt.Sprintf("%s%s", pi.baseURL, g)

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
