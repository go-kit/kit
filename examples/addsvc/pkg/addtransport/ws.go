package addtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-stomp/stomp/frame"
	"github.com/gorilla/websocket"

	"github.com/go-kit/kit/examples/addsvc/pkg/addendpoint"
	"github.com/go-kit/kit/log"
	wstransport "github.com/go-kit/kit/transport/http/ws"
)

func NewWSHandler(endpoints addendpoint.Set, logger log.Logger) http.Handler {
	m := http.NewServeMux()
	m.Handle("/ws", wstransport.NewServer(
		websocket.Upgrader{
			Subprotocols: []string{"v12.stomp"},
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		makeWSSubprotocolCodecMap(),
		makeWSEndpointCodecMap(endpoints),
	))
	return m
}

func makeWSSubprotocolCodecMap() wstransport.SubprotocolCodecMap {
	return wstransport.SubprotocolCodecMap{
		"v12.stomp": wstransport.SubprotocolCodec{
			Decode: decodeStompRequest,
			Encode: encodeStompResponse,
		},
	}
}

// makeEndpointCodecMap returns a codec map configured for the addsvc.
func makeWSEndpointCodecMap(endpoints addendpoint.Set) wstransport.EndpointCodecMap {
	return wstransport.EndpointCodecMap{
		"sum": wstransport.EndpointCodec{
			Endpoint: endpoints.SumEndpoint,
			Decode:   decodeWSSumRequest,
			Encode:   encodeWSSumResponse,
		},
		"concat": wstransport.EndpointCodec{
			Endpoint: endpoints.ConcatEndpoint,
			Decode:   decodeWSConcatRequest,
			Encode:   encodeWSConcatResponse,
		},
	}
}

func decodeStompRequest(_ context.Context, req io.Reader) (string, io.Reader, error) {
	rf := frame.NewReader(req)
	f, err := rf.Read()
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	if f.Command != frame.SEND {
		return "", nil, errors.New("SEND messages supported only")
	}

	return f.Header.Get("destination"), bytes.NewReader(f.Body), nil
}

func encodeStompResponse(_ context.Context, topic string, res io.Writer, msg io.Reader) error {
	body, err := ioutil.ReadAll(msg)
	if err != nil {
		return err
	}

	var h frame.Header
	h.Add("destination", topic)
	h.Add("content-type", "application/json")

	f := frame.Frame{
		Command: frame.MESSAGE,
		Header:  &h,
		Body:    body,
	}

	w := frame.NewWriter(res)
	return w.Write(&f)
}

func decodeWSSumRequest(_ context.Context, msg io.Reader) (interface{}, error) {
	var req addendpoint.SumRequest
	err := json.NewDecoder(msg).Decode(&req)
	return req, err
}

func encodeWSSumResponse(_ context.Context, res io.Writer, msg interface{}) error {
	return json.NewEncoder(res).Encode(msg)
}

func decodeWSConcatRequest(_ context.Context, msg io.Reader) (interface{}, error) {
	var req addendpoint.ConcatRequest
	err := json.NewDecoder(msg).Decode(&req)
	return req, err
}

func encodeWSConcatResponse(_ context.Context, res io.Writer, msg interface{}) error {
	return json.NewEncoder(res).Encode(msg)
}
