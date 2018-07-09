package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var conf *oauth2.Config

type AnnotateFileResponse struct {
	Responses []*AnnotateImageResponse `json:"responses,omitempty"`
}

type AnnotateImageResponse struct {
	LabelAnnotations []*LabelAnnotations `json:"labelAnnotations,omitempty"`

	CustomLabelAnnotations []*CustomLabelAnnotations `json:"customlabelAnnotations,omitempty"`
}

type LabelAnnotations struct {
	Mid         string  `json:"mid,omitempty"`
	Description string  `json:"description,omitempty"`
	Score       float32 `json:"score,omitempty"`
	Topicality  float32 `json:"topicality,omitempty"`
}

func (c LabelAnnotations) String() string {
	return fmt.Sprintf("Label [%s] - [%f]", c.Description, c.Score)
}

type CustomLabelAnnotations struct {
	Model string  `json:"model,omitempty"`
	Label string  `json:"label,omitempty"`
	Score float32 `json:"score,omitempty"`
}

func (c CustomLabelAnnotations) String() string {
	return fmt.Sprintf("Label [%s] - [%f]", c.Label, c.Score)
}

func getClient(serviceAccountEmail, privateKey string) *http.Client {
	conf := &jwt.Config{
		Email:      serviceAccountEmail,
		Scopes:     []string{"https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/cloud-vision"},
		PrivateKey: []byte(privateKey),
		TokenURL:   google.JWTTokenURL,
	}
	// Initiate an http.Client, the following GET request will be
	// authorized and authenticated on the behalf of user@example.com.
	return conf.Client(oauth2.NoContext)
}

func makeRequest(client *http.Client, payload []byte) *AnnotateFileResponse {

	resp, err := client.Post("https://alpha-vision.googleapis.com/v1/images:annotate", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return nil
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return nil
	}

	var afr = new(AnnotateFileResponse)
	err = json.Unmarshal(data, &afr)
	if err != nil {
		fmt.Println("whoops:", err)
	}

	return afr
}

func SendBinaryRequest(client *http.Client, modelURI, inputBase64 string) *AnnotateFileResponse {

	request := []byte(`{
		"requests": [
		  {
			"image": {
				"content":"` + inputBase64 + `"

			},
			"features": [
			  {"type": "CUSTOM_LABEL_DETECTION", "maxResults": 10 },
			  {"type": "LABEL_DETECTION", "maxResults": 10 }
			],
			"customLabelDetectionModels":
				"` + modelURI + `"
		  }
		]
	  }`)

	return makeRequest(client, request)
}

func SendBucketRequest(client *http.Client, modelURI, bucketUrl string) *AnnotateFileResponse {

	request := []byte(`{
		"requests": [
		  {
			"image": {
				"source": {
					"gcsImageUri": "` + bucketUrl + `"
				}
			},
			"features": [
			  {"type": "CUSTOM_LABEL_DETECTION", "maxResults": 10 },
			  {"type": "LABEL_DETECTION", "maxResults": 10 }
			],
			"customLabelDetectionModels":
			"` + modelURI + `"
		  }
		]
	  }`)

	return makeRequest(client, request)
}
