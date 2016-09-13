package goSocketIOClient

import (
	"net/url"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in    []byte
		event string
		data  string
	}{
		{append([]byte{0, 1, 6, 3, 255}, []byte("/message,[\"test\", \"blabla\"]")...), "test", "blabla"},
		{append(append([]byte{0, 1, 6, 3, 255}, []byte("/message,[\"test\", \"blabla\"]")...), append([]byte{0, 1, 6, 3, 255}, []byte("/message,[\"test\", \"blabla\"]")...)...), "test", "blabla"},
		{append(append([]byte{0, 6, 8, 255, 52, 50}, []byte("/message,[\"test\", \"blabla\"]")...), append([]byte{0, 6, 8, 255, 52, 50}, []byte("/message,[\"test\", \"blabla\"]")...)...), "test", "blabla"},
	}
	for _, c := range cases {
		msg := Message{}
		msg.parse(c.in)
		if msg.Event != c.event {
			t.Errorf("Did not get expected result")
		}
	}
}

func TestGetUrl(t *testing.T) {
	u, err := url.Parse("https://bla.net/test")
	if err != nil {
		t.Errorf("Returned error: " + err.Error())
	}

	conn := Conn{SID: "123", Url: u}

	if conn.getURL(false) == "https://bla.net/socket.io/?EIO=3&transport=polling" {
		t.Errorf("Did not get expected result")
	}

	if conn.getURL(true) == "https://bla.net/socket.io/?EIO=3&transport=polling&sid=123" {
		t.Errorf("Did not get expected result")
	}

}
