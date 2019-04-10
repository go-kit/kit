# WebSocket

[JSON RPC](http://www.ws.org) is "A light weight remote procedure call protocol". It allows for the creation of simple RPC-style APIs with human-readable messages that are front-end friendly.

## Using WebSocket with Go-Kit
Using WebSocket and go-kit together is quite simple.

A WebSocket _server_ acts as an [HTTP Handler](https://godoc.org/net/http#Handler), receiving all requests to the WebSockets's URL. On initial handshake to the WebSocket URL the server performs the following:

1. The server upgrades the incoming `http.Request` with the configured `github.com/gorilla/websocket.Upgrader` with the [`Subprotocols`](https://tools.ietf.org/html/rfc6455#section-1.9) you wish to support.
2. Once the Subprotocol requested by the client is verified by the server that it is supported it will fire off a go routine with the WebSocket connection and `SubprotocolCodec`.
3. Each WebSocket message is decoded by the `DecodeSubprotocolFunc` and the resulting _method_ is implemented as an `EndpointCodec`, a go-kit [Endpoint](https://godoc.org/github.com/go-kit/kit/endpoint#Endpoint), sandwiched between a decoder and encoder. The decoder picks apart the message request params, which can be passed to your endpoint as an `io.Reader`.
4. The encoder receives the output from the endpoint and encodes the result to the `io.Writer` which is passed through the `EncodeSubprotocolFunc` as well so it can be sent back to the client.

## Example — Add Service

### `SubprotocolCodecMap`

Let's say we want a service that adds two ints together and utilize the [STOMP](https://stomp.github.io) protocol with JSON messages. We'll serve this at `ws://localhost/ws`. First we establish a WebSocket connection specifying the STOP protocol.

	ws.SubprotocolCodecMap{
		"v12.stomp": ws.SubprotocolCodec{
			Decode: decodeStompRequest,
			Encode: encodeStompResponse,
		},
	}

The `Sec-WebSocket-Protocol` maps to our `SubprotocolCodecMap`.

	GET /ws HTTP/1.1
	Host: localhost
	Upgrade: websocket
	Connection: Upgrade
	Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
	Sec-WebSocket-Protocol: v12.stomp
	Sec-WebSocket-Version: 13
	Origin: http://example.com

Our client connection should receive an acknoledgement of the request.

	HTTP/1.1 101 Switching Protocols
	Upgrade: websocket
	Connection: Upgrade
	Sec-WebSocket-Accept: HSmrc0sMlYUkAGmm5OPpG2HaGWk=
	Sec-WebSocket-Protocol: v12.stomp

So a request to our `sum` endpoint will be STOMP message:

	SEND
	destination:sum
	content-type:application/json

	{"a":1,"b":2}\0

The `destination` header is what our `DecodeSubprotocolFunc` maps to our `EndpointCodecMap`.

### Subprotocol Decoder

	DecodeSubprotocolFunc func(context.Context, io.Reader) (string, io.Reader, error)

A `DecodeSubprotocolFunc` is given the `io.Reader` from the WebSocket connection. It returns an object that will be the input to the Endpoint. For our purposes, the output should be a SumRequest, like this:

	type SumRequest struct {
		A, B int
	}

So here's our decoder:

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

A new `io.Reader` with just the message body will now be passed to the endpoint decoder. Once the endpoint has done its work, we hand over to the…

### Subprotocol Encoder
	
	EncodeSubprotocolFunc func(context.Context, string, io.Writer, io.Reader) error

The encoder takes the output of the endpoint encoder, and writes the message to the `io.Writer` that is sent back to the client in STOMP protocol format. Here's our encoder:

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

### `EndpointCodecMap`
The routing table for incoming requests is the `EndpointCodecMap`. The key of the map is the STOMP destination header value. Here, we're routing the `sum` method to an `EndpointCodec` wrapped around `sumEndpoint`.

	ws.EndpointCodecMap{
		"sum": ws.EndpointCodec{
			Endpoint: sumEndpoint,
			Decode:   decodeSumRequest,
			Encode:   encodeSumResponse,
		},
	}

### Endpoint Decoder
	type DecodeRequestFunc func(context.Context, json.RawMessage) (request interface{}, err error)

A `DecodeRequestFunc` is given the raw JSON from the `params` property of the Request object, _not_ the whole request object. It returns an object that will be the input to the Endpoint. For our purposes, the output should be a SumRequest, like this:

	type SumRequest struct {
		A, B int
	}

So here's our decoder:

	func decodeWSSumRequest(_ context.Context, msg io.Reader) (interface{}, error) {
		var req addendpoint.SumRequest
		err := json.NewDecoder(msg).Decode(&req)
		return req, err
	}

So our `SumRequest` will now be passed to the endpoint. Once the endpoint has done its work, we hand over to the…

### Endpoint Encoder
The encoder takes the output of the endpoint, and builds the raw JSON message that will form the `result` field of a [Response Object](http://www.ws.org/specification#response_object). Our result is going to be a plain int. Here's our encoder:

	func encodeWSSumResponse(_ context.Context, res io.Writer, msg interface{}) error {
		return json.NewEncoder(res).Encode(msg)
	}

### Server
Now that we have an SubprotocolCodec and an EndpointCodec we can wire up the server:

	handler := ws.NewServer(
		websocket.Upgrader{
			Subprotocols: []string{"v12.stomp"},
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		ws.SubprotocolCodecMap{
			"v12.stomp": ws.SubprotocolCodec{
				Decode: decodeStompRequest,
				Encode: encodeStompResponse,
			},
		},
		ws.EndpointCodecMap{
			"sum": ws.EndpointCodec{
				Endpoint: sumEndpoint,
				Decode:   decodeSumRequest,
				Encode:   encodeSumResponse,
			},
		}
	)
	http.Handle("/ws", handler)
	http.ListenAndServe(":80", nil)

With all of this done, our example request above should result in a response like this:

	MESSAGE
	destination:sum
	content-type:application/json

	{"v":3}\0
