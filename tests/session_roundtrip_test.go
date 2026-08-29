package tests

import (
	"bufio"
	"bytes"
	"net/http"
	"reflect"
	"testing"
	"time"

	"replaynet/internal/session"
)

func TestSessionEventRoundTrip(t *testing.T) {
	cases := []session.Event{
		{
			Offset:     0,
			Type:       session.EventRequest,
			Method:     "GET",
			Path:       "/",
			StatusCode: 0,
			Headers:    http.Header{},
			Body:       nil,
		},
		{
			Offset:     42 * time.Millisecond,
			Type:       session.EventResponse,
			Method:     "",
			Path:       "",
			StatusCode: 200,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"ok":true}`),
		},
		{
			Offset:     100 * time.Millisecond,
			Type:       session.EventRequest,
			Method:     "POST",
			Path:       "/upload/日本語",
			StatusCode: 0,
			Headers: http.Header{
				"X-Multi": []string{"a", "b", "c"},
				"Empty":   []string{""},
			},
			Body: bytes.Repeat([]byte{0xAB}, 2*1024*1024),
		},
		{
			Offset:     0,
			Type:       session.EventResponse,
			Method:     "",
			Path:       "",
			StatusCode: 500,
			Headers:    http.Header{},
			Body:       []byte{},
		},
	}

	for i, want := range cases {
		buf := &bytes.Buffer{}
		if err := session.WriteEvent(buf, want); err != nil {
			t.Fatalf("case %d: write error: %v", i, err)
		}

		got, err := session.ReadEvent(bufio.NewReader(buf))
		if err != nil {
			t.Fatalf("case %d: read error: %v", i, err)
		}

		if got.Offset != want.Offset {
			t.Errorf("case %d: offset mismatch: got %v want %v", i, got.Offset, want.Offset)
		}
		if got.Type != want.Type {
			t.Errorf("case %d: type mismatch: got %v want %v", i, got.Type, want.Type)
		}
		if got.Method != want.Method {
			t.Errorf("case %d: method mismatch: got %q want %q", i, got.Method, want.Method)
		}
		if got.Path != want.Path {
			t.Errorf("case %d: path mismatch: got %q want %q", i, got.Path, want.Path)
		}
		if got.StatusCode != want.StatusCode {
			t.Errorf("case %d: status mismatch: got %d want %d", i, got.StatusCode, want.StatusCode)
		}
		if !bytes.Equal(got.Body, want.Body) {
			t.Errorf("case %d: body mismatch: got len %d want len %d", i, len(got.Body), len(want.Body))
		}
		if !reflect.DeepEqual(got.Headers, want.Headers) {
			t.Errorf("case %d: headers mismatch: got %v want %v", i, got.Headers, want.Headers)
		}
	}
}

func TestSessionWriterLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.rnet"

	w, err := session.NewWriter(path)
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}

	events := []session.Event{
		{Type: session.EventRequest, Method: "GET", Path: "/a", Headers: http.Header{}},
		{Type: session.EventResponse, StatusCode: 200, Headers: http.Header{}, Body: []byte("hello")},
		{Type: session.EventRequest, Method: "GET", Path: "/a", Headers: http.Header{}},
		{Type: session.EventResponse, StatusCode: 500, Headers: http.Header{}, Body: []byte("err")},
	}

	for _, e := range events {
		if err := w.Append(e); err != nil {
			t.Fatalf("append: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	sess, err := session.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(sess.Events) != len(events) {
		t.Fatalf("event count mismatch: got %d want %d", len(sess.Events), len(events))
	}

	for i, want := range events {
		got := sess.Events[i]
		if got.Index != i {
			t.Errorf("event %d: index mismatch: got %d want %d", i, got.Index, i)
		}
		if got.Method != want.Method || got.Path != want.Path || got.StatusCode != want.StatusCode {
			t.Errorf("event %d: mismatch: got %+v want %+v", i, got, want)
		}
		if !bytes.Equal(got.Body, want.Body) {
			t.Errorf("event %d: body mismatch", i)
		}
	}
}
