package edgemax

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// A Client is a client for a Ubiquiti EdgeMAX device.
//
// Client.Login must be called and return a nil error before any additional
// actions can be performed with a Client.
type Client struct {
	client *http.Client
	url    *url.URL
}

const (
	// userAgent is the default user agent this package will report to
	// the EdgeMAX device.
	userAgent = "github.com/vaga/edgemax_exporter"
	// sessionCookie is the name of the session cookie used to authenticate
	// against EdgeMAX devices.
	sessionCookie = "PHPSESSID"
)

// NewClient creates a new Client, using the input EdgeMAX device address
// and an optional HTTP client. If no HTTP client is specified, a default
// one will be used.
//
// Client.Login must be called and return a nil error before any additional
// actions can be performed with a Client.
func NewClient(addr string, client *http.Client) (*Client, error) {
	u, err := url.Parse(strings.TrimRight(addr, "/"))
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	if client.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		client.Jar = jar
	}

	return &Client{
		client: client,
		url:    u,
	}, nil
}

// Login authenticates against the EdgeMAX device using the specified username
// and password. Login must be called and return a nil error before any
// additional actions can be performed.
func (c *Client) Login(username, password string) error {
	v := make(url.Values, 2)
	v.Set("username", username)
	v.Set("password", password)

	_, err := c.client.PostForm(c.url.String(), v)
	return err
}

// Stats opens a websocket connection to an EdgeMAX device to retrieve
// statistics which are sent using the socket.
func (c *Client) Stats(
	systemCh chan<- SystemStat,
	dpiCh chan<- DPIStat,
	ifacesCh chan<- InterfacesStat,
) (func(), error) {
	conn, err := c.dial()
	if err != nil {
		return nil, err
	}

	var sessionID string
	for _, c := range c.client.Jar.Cookies(c.url) {
		if c.Name == sessionCookie {
			sessionID = c.Value
			break
		}
	}

	if err := conn.WriteMessage(websocket.TextMessage, marshalWS(
		connectRequest{
			Subscribe: []stat{
				stat{Name: "system-stats"},
				stat{Name: "export"},
				stat{Name: "interfaces"},
			},
			SessionID: sessionID,
		},
	)); err != nil {
		return nil, err
	}

	doneCh := make(chan struct{})
	wg := new(sync.WaitGroup)

	go c.keepAlive(wg, doneCh)
	go c.read(conn, wg, doneCh, systemCh, dpiCh, ifacesCh)

	return func() { close(doneCh); wg.Wait() }, nil
}

// dial initializes the websocket used for Client.Stats
func (c *Client) dial() (*websocket.Conn, error) {
	// Websocket URL is adapted from HTTP URL
	wsURL := *c.url
	wsURL.Scheme = "wss"
	wsURL.Path = "/ws/stats"

	d := &websocket.Dialer{EnableCompression: true}

	// Copy TLS config from client if using standard *http.Transport, so that
	// using InsecureHTTPClient can also apply to websocket connections
	if tr, ok := c.client.Transport.(*http.Transport); ok {
		d.TLSClientConfig = tr.TLSClientConfig
	}

	conn, _, err := d.Dial(wsURL.String(), http.Header{
		"Origin": []string{c.url.Scheme + "://" + c.url.Host},
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// read receives raw stats from a websocket and decodes them into
// Stat structs of various types.
func (c *Client) read(
	conn *websocket.Conn,
	wg *sync.WaitGroup,
	doneCh chan struct{},
	systemCh chan<- SystemStat, dpiCh chan<- DPIStat, ifacesCh chan<- InterfacesStat,
) {
	defer conn.Close()
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-doneCh:
			return
		default:
		}

		_, m, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			close(doneCh)
			return
		}
		rm := make(map[string]json.RawMessage)
		if err := unmarshalWS(m, &rm); err != nil {
			log.Println("unmarshal:", err)
		}

		for sn, sk := range rm {
			switch sn {
			case "system-stats":
				var s SystemStat
				if err := json.Unmarshal(sk, &s); err != nil {
					log.Println("unmarshal system-stats:", err)
				}
				systemCh <- s
			case "export":
				var s DPIStat
				if err := json.Unmarshal(sk, &s); err != nil {
					log.Println("unmarshal export:", err)
				}
				dpiCh <- s
			case "interfaces":
				var s InterfacesStat
				if err := json.Unmarshal(sk, &s); err != nil {
					log.Println("unmarshal interfaces:", err)
				}
				ifacesCh <- s
			}
		}
	}
}

// keepalive sends heartbeat requests at regular intervals to the EdgeMAX
// device to keep a session active while Client.Stats is running.
func (c *Client) keepAlive(wg *sync.WaitGroup, doneCh chan struct{}) {
	wg.Add(1)
	defer wg.Done()

	for {
		_, err := c.client.Get(fmt.Sprintf("%s/api/edge/heartbeat.json?_=%d", c.url.String(), time.Now().UnixNano()))
		if err != nil {
			log.Printf("could not request edgemax API: %v", err)
		}
		select {
		case <-time.After(10 * time.Second):
		case <-doneCh:
			return
		}
	}
}
