package goSocketIOClient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	URL "net/url"
)

// Conn connection information
type Conn struct {
	SID       string
	Url       *URL.URL
	Transport string
	Output    chan Message
}

// Message reviced from server
type Message struct {
	Event string
	Data  string
}

type errorMsg struct {
	Code    int    `json: "code"`
	Message string `json: "message"`
}

func (msg *Message) parse(b []byte) ([]byte, error) {
	var rest []byte
	if bytes.Index(b, []byte("[")) < 1 {
		msgData := errorMsg{}
		json.Unmarshal(b, &msgData)
		if msgData.Message == "Session ID unknown" {
			return nil, errors.New("Session ID unknown")
		}
		return nil, errors.New("Parsing error")
	}

	cut := b[bytes.Index(b, []byte("[")):]

	if bytes.Index(cut, []byte{255}) > 0 { // Are there more than one msg
		rest = cut[bytes.Index(cut, []byte{255}):]
		cut = cut[:bytes.Index(cut, []byte{255})]
		cut = cut[:bytes.LastIndex(cut, []byte("]"))+1]
	}

	var data []string
	if err := json.Unmarshal(cut, &data); err != nil {
		fmt.Println(b)
		return nil, err
	}

	msg.Event = data[0]
	msg.Data = data[1]

	return rest, nil
}

func (conn *Conn) getData() ([]Message, error) {
	var messages []Message

	resp, err := http.Get(conn.getURL(true))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	for { // Somethimes there are more then one
		msg := Message{}
		rest, err := msg.parse(body)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
		if rest == nil {
			break
		}
		body = rest
	}

	return messages, nil
}

func (conn *Conn) start() {
	go func() {
		for {
			msgs, err := conn.getData()
			if err != nil {
				if err.Error() == "Session ID unknown" {
					log.Println("Session ID unknown. Preforming new handshake.")
					err = conn.pollingHandshake()
					if err != nil {
						log.Fatal(err)
					}
				} else {
					log.Println(err)
				}

				continue
			}
			for _, msg := range msgs {
				conn.Output <- msg
			}
		}
	}()
}

func (conn *Conn) pollingHandshake() error {
	resp, err := http.Get(conn.getURL(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slice := make([]byte, 5) // Remove the first 5 bytes
	resp.Body.Read(slice)

	t := &struct {
		SID string `json: "sid"`
	}{}
	json.NewDecoder(resp.Body).Decode(t)

	postResp, err := http.Post(conn.getURL(false)+"&sid="+t.SID,
		"text/html", bytes.NewBuffer([]byte("10:40"+conn.Url.Path)))
	if err != nil {
		return err
	}

	defer postResp.Body.Close()
	body, err := ioutil.ReadAll(postResp.Body)
	if err != nil {
		return err
	}

	if (string(body)) != "ok" {
		return errors.New("Handshake did not return ok")
	}

	conn.SID = t.SID

	return nil
}

func (conn *Conn) getURL(withSID bool) string {
	url := conn.Url.Scheme + "://" + conn.Url.Host + "/socket.io/?EIO=3&transport=" + conn.Transport
	if withSID {
		url = url + "&sid=" + conn.SID
	}
	return url
}

// Dial connect to a socket.io host
func Dial(url string) (Conn, error) {
	u, err := URL.Parse(url)
	if err != nil {
		return Conn{}, err
	}

	conn := Conn{
		Url:       u,
		Transport: "polling",
	}

	err = conn.pollingHandshake()
	if err != nil {
		log.Fatal(err)
	}

	conn.Output = make(chan Message, 10)

	conn.start()

	return conn, err
}
