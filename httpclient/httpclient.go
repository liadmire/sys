package httpclient

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type IHttpClient interface {
	Get()
	Post()
	Put()
	Delete()
	Head()
}

type ContentType int

const (
	ContentTypeForm ContentType = iota
	ContentTypeJSON
)

type HTTPSettings struct {
	ShowDebug        bool
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	TLSClientConfig  *tls.Config
	Proxy            func(*http.Request) (*url.URL, error)
	Transport        http.RoundTripper
	CheckRedirect    func(req *http.Request, via []*http.Request) error
	EnableCookie     bool
	Gzip             bool
	DumpBody         bool
	Retries          int
}

type THttpClient struct {
	baseURL     string
	relativeURL string
	contentType ContentType
	req         *http.Request
	params      map[string]interface{}
	files       map[string]string
	setting     HTTPSettings
	resp        *http.Response
	body        []byte
	dump        []byte
}

var defaultSetting = HTTPSettings{
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
	Gzip:             true,
	DumpBody:         true,
}

var defaultCookieJar http.CookieJar
var settingMutex sync.Mutex

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultCookieJar, _ = cookiejar.New(nil)
}

// SetDefaultSetting Overwrite default settings
func SetDefaultSetting(setting HTTPSettings) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting = setting
}

func NewHttpClient(baseURL string) *THttpClient {
	var resp http.Response
	u, err := url.Parse(baseURL)
	if err != nil {
		log.Println("httpclient:", err)
	}
	req := http.Request{
		URL:        u,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &THttpClient{
		baseURL: baseURL,
		req:     &req,
		params:  map[string]interface{}{},
		files:   map[string]string{},
		setting: defaultSetting,
		resp:    &resp,
	}
}

func (hc *THttpClient) GetRequest() *http.Request {
	return hc.req
}

func (hc *THttpClient) SetBasicAuth(username, password string) *THttpClient {
	hc.req.SetBasicAuth(username, password)
	return hc
}

func (hc *THttpClient) Header(key, value string) *THttpClient {
	hc.req.Header.Set(key, value)
	return hc
}

func (hc *THttpClient) SetHost(host string) *THttpClient {
	hc.req.Host = host
	return hc
}

func (hc *THttpClient) SetContentType(contentType ContentType) *THttpClient {
	hc.contentType = contentType
	return hc
}

func (hc *THttpClient) Response() (*http.Response, error) {
	return hc.getResponse()
}

func (hc *THttpClient) getResponse() (*http.Response, error) {
	if hc.resp.StatusCode != 0 {
		return hc.resp, nil
	}
	resp, err := hc.DoRequest()
	if err != nil {
		return nil, err
	}
	hc.resp = resp
	return resp, nil
}

func (hc *THttpClient) DoRequest() (resp *http.Response, err error) {
	var paramBody string
	if len(hc.params) > 0 {
		switch hc.contentType {
		case ContentTypeForm:
			var buf bytes.Buffer
			for k, v := range hc.params {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				if vv, ok := v.(string); ok {
					buf.WriteString(url.QueryEscape(vv))
					buf.WriteByte('&')
				}
			}
			paramBody = buf.String()
			paramBody = paramBody[0 : len(paramBody)-1]

		case ContentTypeJSON:
			data := make(map[string]interface{})
			for k, v := range hc.params {
				data[k] = v
			}

			pbody, err := json.Marshal(data)
			if err != nil {
				fmt.Println("param body data err： ", err)
				return nil, err
			}

			paramBody = string(pbody)
		}
	}

	hc.buildURL(paramBody)
	urlParsed, err := url.Parse(fmt.Sprintf("%s%s", hc.baseURL, hc.relativeURL))
	if err != nil {
		return nil, err
	}

	hc.req.URL = urlParsed

	trans := hc.setting.Transport

	if trans == nil {
		// create default transport
		trans = &http.Transport{
			TLSClientConfig:     hc.setting.TLSClientConfig,
			Proxy:               hc.setting.Proxy,
			Dial:                TimeoutDialer(hc.setting.ConnectTimeout, hc.setting.ReadWriteTimeout),
			MaxIdleConnsPerHost: 100,
		}
	} else {
		// if hc.transport is *http.Transport then set the settings.
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = hc.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = hc.setting.Proxy
			}
			if t.Dial == nil {
				t.Dial = TimeoutDialer(hc.setting.ConnectTimeout, hc.setting.ReadWriteTimeout)
			}
		}
	}

	var jar http.CookieJar
	if hc.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}

	if hc.setting.UserAgent != "" && hc.req.Header.Get("User-Agent") == "" {
		hc.req.Header.Set("User-Agent", hc.setting.UserAgent)
	}

	if hc.setting.CheckRedirect != nil {
		client.CheckRedirect = hc.setting.CheckRedirect
	}

	if hc.setting.ShowDebug {
		dump, err := httputil.DumpRequest(hc.req, hc.setting.DumpBody)
		if err != nil {
			log.Println(err.Error())
		}
		hc.dump = dump
	}
	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries is setted, it will retries fixed times.
	for i := 0; hc.setting.Retries == -1 || i <= hc.setting.Retries; i++ {
		resp, err = client.Do(hc.req)
		if err == nil {
			break
		}
	}
	return resp, err
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}

func (hc *THttpClient) buildURL(paramBody string) {
	// build GET url with query string
	if hc.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(hc.relativeURL, "?") {
			hc.relativeURL += "&" + paramBody
		} else {
			hc.relativeURL = hc.relativeURL + "?" + paramBody
		}
		return
	}

	// build POST/PUT/PATCH url and body
	if (hc.req.Method == "POST" || hc.req.Method == "PUT" || hc.req.Method == "PATCH" || hc.req.Method == "DELETE") && hc.req.Body == nil {
		// with files
		if len(hc.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formname, filename := range hc.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					if err != nil {
						log.Println("Httplib:", err)
					}
					fh, err := os.Open(filename)
					if err != nil {
						log.Println("Httplib:", err)
					}
					//iocopy
					_, err = io.Copy(fileWriter, fh)
					fh.Close()
					if err != nil {
						log.Println("Httplib:", err)
					}
				}
				for k, v := range hc.params {
					vv, ok := v.(string)
					if ok {
						bodyWriter.WriteField(k, vv)
					}
				}
				bodyWriter.Close()
				pw.Close()
			}()
			hc.Header("Content-Type", bodyWriter.FormDataContentType())
			hc.req.Body = ioutil.NopCloser(pr)
			return
		}

		// with params
		if len(paramBody) > 0 {
			switch hc.contentType {
			case ContentTypeForm:
				hc.Header("Content-Type", "application/x-www-form-urlencoded")
			case ContentTypeJSON:
				hc.Header("Content-Type", "application/json")
			default:
				hc.Header("Content-Type", "text/plain")
			}
			hc.Body(paramBody)
		}
	}
}

func (hc *THttpClient) Body(data interface{}) *THttpClient {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		hc.req.Body = ioutil.NopCloser(bf)
		hc.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		hc.req.Body = ioutil.NopCloser(bf)
		hc.req.ContentLength = int64(len(t))
	}
	return hc
}

// Bytes x
func (hc *THttpClient) Bytes() ([]byte, error) {
	if hc.body != nil {
		return hc.body, nil
	}
	resp, err := hc.getResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if hc.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		hc.body, err = ioutil.ReadAll(reader)
		return hc.body, err
	}
	hc.body, err = ioutil.ReadAll(resp.Body)
	return hc.body, err
}

// ToJSON x
func (hc *THttpClient) ToJSON(v interface{}) error {
	data, err := hc.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML x
func (hc *THttpClient) ToXML(v interface{}) error {
	data, err := hc.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToFile x
func (hc *THttpClient) ToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := hc.getResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func (hc *THttpClient) Param(key string, value interface{}) *THttpClient {
	if _, ok := hc.params[key]; ok {
		hc.params[key] = value
		fmt.Printf("key %s exists!", key)
	} else {
		hc.params[key] = value
	}
	return hc
}

func (hc *THttpClient) Get(rURL string) *THttpClient {
	hc.relativeURL = rURL
	hc.req.Method = "GET"

	return hc
}

func (hc *THttpClient) Post(rURL string) *THttpClient {
	hc.relativeURL = rURL
	hc.req.Method = "POST"
	return hc
}

func (hc *THttpClient) Put(rURL string) *THttpClient {
	hc.relativeURL = rURL
	hc.req.Method = "PUT"
	return hc
}

func (hc *THttpClient) Delete(rURL string) *THttpClient {
	hc.relativeURL = rURL
	hc.req.Method = "DELETE"
	return hc
}

func (hc *THttpClient) Head(rURL string) *THttpClient {
	hc.relativeURL = rURL
	hc.req.Method = "HEAD"
	return hc
}

// Get returns *THttpClient with GET method.
func Get(url string) *THttpClient {
	return NewHttpClient(url).Get("")
}

// Post returns *THttpClient with POST method.
func Post(url string) *THttpClient {
	return NewHttpClient(url).Post("")
}

// Put returns *THttpClient with PUT method.
func Put(url string) *THttpClient {
	return NewHttpClient(url).Put("")
}

// Delete returns *THttpClient DELETE method.
func Delete(url string) *THttpClient {
	return NewHttpClient(url).Delete("")
}

// Head returns *THttpClient with HEAD method.
func Head(url string) *THttpClient {
	return NewHttpClient(url).Head("")
}
