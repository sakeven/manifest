package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Client client
type Client struct {
	client   *http.Client
	APIPath  string
	Username string
	Password string
}

var defaultTimeout = 600 * time.Second

// NewClient creates a new Registry client with a default timeout.
func NewClient(apiPath, username, password string) *Client {
	return NewClientTimeout(apiPath, username, password, defaultTimeout)
}

// NewClientTimeout acts like NewClient but takes a timeout.
func NewClientTimeout(apiPath, username, password string, timeout time.Duration) *Client {
	if apiPath == "docker.io" {
		apiPath = "https://registry-1.docker.io"
	}

	apiPath = strings.TrimRight(apiPath, "/")
	if !strings.HasPrefix(apiPath, "http") {
		apiPath = "http://" + apiPath
	}

	return &Client{
		client:   &http.Client{Timeout: timeout},
		APIPath:  apiPath,
		Username: username,
		Password: password,
	}
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (r *Client) Do(req *http.Request) (*http.Response, error) {
	return r.client.Do(req)
}

// do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (r *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c := resp.StatusCode; !(200 <= c && c <= 299) {
		content, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %d, body %s", c, string(content))
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	return resp, err
}

// newRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.  If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (r *Client) newRequest(method string, url string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(url, "http") {
		url = r.APIPath + url
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	token, err := r.getToken(method, url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return req, nil
}

// getToken gets a token for auth
func (r *Client) getToken(method, url string) (string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get www-authenticate failed %s", err)
	}
	defer resp.Body.Close()

	wwwAuthenticate := resp.Header.Get("Www-Authenticate")
	var realm, service, scope string
	fmt.Sscanf(wwwAuthenticate, `Bearer realm=%q,service=%q,scope=%q`, &realm, &service, &scope)

	authURL := fmt.Sprintf("%s?service=%s&scope=%s", realm, service, scope)
	req, err = http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}
	if len(r.Username) > 0 {
		req.SetBasicAuth(r.Username, r.Password)
	}

	token := struct {
		Token string `json:"token"`
	}{}

	_, err = r.do(req, &token)
	return token.Token, err
}
