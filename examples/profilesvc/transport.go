package main

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

var (
	errBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func makeHandler(ctx context.Context, s ProfileService, logger kitlog.Logger) stdhttp.Handler {
	e := makeEndpoints(s)
	r := mux.NewRouter()

	commonOptions := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	// POST    /profiles                           adds another profile
	// GET     /profiles/:id                       retrieves the given profile by id
	// PUT     /profiles/:id                       post updated profile information about the profile
	// PATCH   /profiles/:id                       partial updated profile information
	// DELETE  /profiles/:id                       remove the given profile
	// GET     /profiles/:id/addresses             retrieve addresses associated with the profile
	// GET     /profiles/:id/addresses/:addressID  retrieve a particular profile address
	// POST    /profiles/:id/addresses             add a new address
	// DELETE  /profiles/:id/addresses/:addressID  remove an address

	r.Methods("POST").Path("/profiles/").Handler(kithttp.NewServer(
		ctx,
		e.postProfileEndpoint,
		decodePostProfileRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("GET").Path("/profiles/{id}").Handler(kithttp.NewServer(
		ctx,
		e.getProfileEndpoint,
		decodeGetProfileRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("PUT").Path("/profiles/{id}").Handler(kithttp.NewServer(
		ctx,
		e.putProfileEndpoint,
		decodePutProfileRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("PATCH").Path("/profiles/{id}").Handler(kithttp.NewServer(
		ctx,
		e.patchProfileEndpoint,
		decodePatchProfileRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("DELETE").Path("/profiles/{id}").Handler(kithttp.NewServer(
		ctx,
		e.deleteProfileEndpoint,
		decodeDeleteProfileRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("GET").Path("/profiles/{id}/addresses/").Handler(kithttp.NewServer(
		ctx,
		e.getAddressesEndpoint,
		decodeGetAddressesRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("GET").Path("/profiles/{id}/addresses/{addressID}").Handler(kithttp.NewServer(
		ctx,
		e.getAddressEndpoint,
		decodeGetAddressRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("POST").Path("/profiles/{id}/addresses/").Handler(kithttp.NewServer(
		ctx,
		e.postAddressEndpoint,
		decodePostAddressRequest,
		encodeResponse,
		commonOptions...,
	))
	r.Methods("DELETE").Path("/profiles/{id}/addresses/{addressID}").Handler(kithttp.NewServer(
		ctx,
		e.deleteAddressEndpoint,
		decodeDeleteAddressRequest,
		encodeResponse,
		commonOptions...,
	))
	return r
}

func decodePostProfileRequest(r *stdhttp.Request) (request interface{}, err error) {
	var req postProfileRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Profile); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeGetProfileRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	return getProfileRequest{ID: id}, nil
}

func decodePutProfileRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return putProfileRequest{
		ID:      id,
		Profile: profile,
	}, nil
}

func decodePatchProfileRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return patchProfileRequest{
		ID:      id,
		Profile: profile,
	}, nil
}

func decodeDeleteProfileRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	return deleteProfileRequest{ID: id}, nil
}

func decodeGetAddressesRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	return getAddressesRequest{ProfileID: id}, nil
}

func decodeGetAddressRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	addressID, ok := vars["addressID"]
	if !ok {
		return nil, errBadRouting
	}
	return getAddressRequest{
		ProfileID: id,
		AddressID: addressID,
	}, nil
}

func decodePostAddressRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	var address Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		return nil, err
	}
	return postAddressRequest{
		ProfileID: id,
		Address:   address,
	}, nil
}

func decodeDeleteAddressRequest(r *stdhttp.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRouting
	}
	addressID, ok := vars["addressID"]
	if !ok {
		return nil, errBadRouting
	}
	return deleteAddressRequest{
		ProfileID: id,
		AddressID: addressID,
	}, nil
}

// errorer is implemented by all concrete response types. It allows us to
// change the HTTP response code without needing to trigger an endpoint
// (transport-level) error. For more information, read the big comment in
// endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because I didn't know if something more
// specific was necessary. It's certainly possible to specialize on a
// per-response (per-method) basis.
func encodeResponse(w stdhttp.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(w, e.error())
		return nil
	}
	return json.NewEncoder(w).Encode(response)
}

func encodeError(w stdhttp.ResponseWriter, err error) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case errNotFound:
		return stdhttp.StatusNotFound
	case errAlreadyExists, errInconsistentIDs:
		return stdhttp.StatusBadRequest
	default:
		if _, ok := err.(kithttp.BadRequestError); ok {
			return stdhttp.StatusBadRequest
		}
		return stdhttp.StatusInternalServerError
	}
}
