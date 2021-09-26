package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type EsaPostBody struct {
	Name     string `json:"name"`
	BodyMd   string `json:"body_md"`
	Category string `json:"category"`
	WIP      bool   `json:"wip"`
}

type EsaPatchBody struct {
	BodyMd string `json:"body_md"`
}

type EsaResponseBody struct {
	Name     string `json:"name"`
	Number   int    `json:"number"`
	BodyMd   string `json:"body_md"`
	FullName string `json:"full_name"`
	Url      string `json:"url"`
}

func CreateDoc(accessToken string, team string, body EsaPostBody) (*EsaResponseBody, error) {
	URL := fmt.Sprintf("https://api.esa.io/v1/teams/%s/posts", team)

	jsonBody, err := json.Marshal(body)
	fatalIfError(err)

	auth := fmt.Sprintf("Bearer %s", accessToken)
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)

	client := new(http.Client)
	res, err := client.Do(req)
	fatalIfError(err)

	defer res.Body.Close()

	dump, err := httputil.DumpResponse(res, true)
	fatalIfError(err)
	fmt.Println(string(dump))

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	result := EsaResponseBody{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func PatchDoc(accessToken string, team string, postNumber int, body EsaPatchBody) (*EsaResponseBody, error) {
	URL := fmt.Sprintf("https://api.esa.io/v1/teams/%s/posts/%d", team, postNumber)
	jsonBody, err := json.Marshal(body)
	fatalIfError(err)

	auth := fmt.Sprintf("Bearer %s", accessToken)
	req, err := http.NewRequest("PATCH", URL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)

	client := new(http.Client)
	res, err := client.Do(req)
	fatalIfError(err)

	defer res.Body.Close()

	dump, err := httputil.DumpResponse(res, true)
	fatalIfError(err)
	fmt.Println(string(dump))

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	result := EsaResponseBody{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
