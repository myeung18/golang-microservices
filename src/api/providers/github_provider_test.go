package githubprovider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/SophieDeBenedetto/golang-microservices/src/api/clients/restclient"
	"github.com/SophieDeBenedetto/golang-microservices/src/api/domain/github"

	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return getDoFunc(req)
}

var (
	// mockClient MockClient
	getDoFunc func(req *http.Request) (*http.Response, error)
)

func init() {
	restclient.Client = &MockClient{}
}

func TestConstants(t *testing.T) {
	assert.EqualValues(t, "Authorization", headerAuthorization)
	assert.EqualValues(t, "token %s", headerAuthorizationFormat)
	assert.EqualValues(t, "https://api.github.com/user/repos", createRepoURL)
}
func TestGetAuthorizationHeader(t *testing.T) {
	accessToken := "ABC123"
	assert.EqualValues(t, "token ABC123", getAuthorizationHeader(accessToken))
}

func TestCreateRepoInvalidRestclientResponse(t *testing.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Invalid token"}`)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       r,
		}, nil
	}
	response, err := CreateRepo("invalidToken", &github.CreateRepoRequest{})
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.EqualValues(t, err.Message, "Invalid token")
}

func TestCreateRepoInvalidResponseBody(t *testing.T) {
	r, _ := os.Open("-asdf123")
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			Body: r,
		}, nil
	}
	response, err := CreateRepo("validToken", &github.CreateRepoRequest{})
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.EqualValues(t, err.Message, "Invalid response body")
}

func TestCreateRepoResponseWithError(t *testing.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Not Found"}`)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       r,
		}, nil
	}
	response, err := CreateRepo("validToken", &github.CreateRepoRequest{})
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.EqualValues(t, "Not Found", err.Message)
}

func TestCreateRepoSuccess(t *testing.T) {
	json := `{"name":"Test Name","full_name":"test full name","owner":{"login": "octocat"}}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	response, err := CreateRepo("validToken", &github.CreateRepoRequest{})
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assert.EqualValues(t, "Test Name", response.Name)
	assert.EqualValues(t, "test full name", response.FullName)
	assert.EqualValues(t, "octocat", response.Owner.Login)
}

func TestCreateRepoInvalidResponseInterface(t *testing.T) {
	fmt.Println("HERE")
	json := `{"id":"sophie"}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	response, err := CreateRepo("validToken", &github.CreateRepoRequest{})
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.EqualValues(t, http.StatusOK, err.StatusCode)
	assert.EqualValues(t, "Error trying to decode successfull create repo JSON response body", err.Message)
}
