package edgemax

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestMarshalWS(t *testing.T) {
	var tests = []struct {
		desc string
		cr   connectRequest
		out  []byte
	}{
		{
			desc: "empty request",
			cr:   connectRequest{},
			out:  append([]byte("53\n"), `{"SUBSCRIBE":null,"UNSUBSCRIBE":null,"SESSION_ID":""}`...),
		},
		{
			desc: "subscribe to two streams",
			cr: connectRequest{
				Subscribe: []stat{
					{Name: "foo"},
					{Name: "bar"},
				},
				SessionID: "baz",
			},
			out: append([]byte("83\n"), `{"SUBSCRIBE":[{"name":"foo"},{"name":"bar"}],"UNSUBSCRIBE":null,"SESSION_ID":"baz"}`...),
		},
		{
			desc: "unsubscribe from one stream",
			cr: connectRequest{
				Unsubscribe: []stat{
					{Name: "foo"},
				},
				SessionID: "bar",
			},
			out: append([]byte("68\n"), `{"SUBSCRIBE":null,"UNSUBSCRIBE":[{"name":"foo"}],"SESSION_ID":"bar"}`...),
		},
	}

	for i, tt := range tests {
		t.Logf("[%02d] test %q", i, tt.desc)

		out := marshalWS(tt.cr)
		if want, got := tt.out, out; !bytes.Equal(want, got) {
			t.Fatalf("unexpected output:\n- want: %q\n-  got: %q", string(want), string(got))
		}
	}
}

func TestUnmarshalWS(t *testing.T) {
	var tests = []struct {
		desc string
		in   []byte
		cr   connectRequest
		err  error
	}{
		{
			desc: "incorrect number of newlines",
			in:   []byte("foo"),
			err:  errors.New("incorrect number of elements in websocket message: 1"),
		},
		{
			desc: "no JSON object present",
			in:   []byte("3\n"),
			cr:   connectRequest{},
		},
		{
			desc: "empty request with no length",
			in:   []byte(`{"SUBSCRIBE":null,"UNSUBSCRIBE":null,"SESSION_ID":""}`),
			cr:   connectRequest{},
		},
		{
			desc: "empty request",
			in:   append([]byte("53\n"), `{"SUBSCRIBE":null,"UNSUBSCRIBE":null,"SESSION_ID":""}`...),
			cr:   connectRequest{},
		},
		{
			desc: "subscribe to two streams",
			in:   append([]byte("83\n"), `{"SUBSCRIBE":[{"name":"foo"},{"name":"bar"}],"UNSUBSCRIBE":null,"SESSION_ID":"baz"}`...),
			cr: connectRequest{
				Subscribe: []stat{
					{Name: "foo"},
					{Name: "bar"},
				},
				SessionID: "baz",
			},
		},
		{
			desc: "unsubscribe from one stream",
			in:   append([]byte("68\n"), `{"SUBSCRIBE":null,"UNSUBSCRIBE":[{"name":"foo"}],"SESSION_ID":"bar"}`...),
			cr: connectRequest{
				Unsubscribe: []stat{
					{Name: "foo"},
				},
				SessionID: "bar",
			},
		},
	}

	for i, tt := range tests {
		t.Logf("[%02d] test %q", i, tt.desc)

		var cr connectRequest
		err := unmarshalWS(tt.in, &cr)
		if want, got := errStr(tt.err), errStr(err); want != got {
			t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
		}
		if err != nil {
			continue
		}

		if want, got := tt.cr, cr; !reflect.DeepEqual(want, got) {
			t.Fatalf("unexpected wsRequest object:\n- want: %v\n-  got: %v", want, got)
		}
	}
}

func errStr(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
