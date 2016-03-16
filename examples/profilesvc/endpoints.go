package main

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

type endpoints struct {
	postProfileEndpoint   endpoint.Endpoint
	getProfileEndpoint    endpoint.Endpoint
	putProfileEndpoint    endpoint.Endpoint
	patchProfileEndpoint  endpoint.Endpoint
	deleteProfileEndpoint endpoint.Endpoint
	getAddressesEndpoint  endpoint.Endpoint
	getAddressEndpoint    endpoint.Endpoint
	postAddressEndpoint   endpoint.Endpoint
	deleteAddressEndpoint endpoint.Endpoint
}

func makeEndpoints(s ProfileService) endpoints {
	return endpoints{
		postProfileEndpoint:   makePostProfileEndpoint(s),
		getProfileEndpoint:    makeGetProfileEndpoint(s),
		putProfileEndpoint:    makePutProfileEndpoint(s),
		patchProfileEndpoint:  makePatchProfileEndpoint(s),
		deleteProfileEndpoint: makeDeleteProfileEndpoint(s),
		getAddressesEndpoint:  makeGetAddressesEndpoint(s),
		getAddressEndpoint:    makeGetAddressEndpoint(s),
		postAddressEndpoint:   makePostAddressEndpoint(s),
		deleteAddressEndpoint: makeDeleteAddressEndpoint(s),
	}
}

type postProfileRequest struct {
	Profile Profile
}

type postProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r postProfileResponse) error() error { return r.Err }

// Regarding errors returned from service (business logic) methods, we have two
// options. We could return the error via the endpoint itself. That makes
// certain things a little bit easier, like providing non-200 HTTP responses to
// the client. But Go kit assumes that endpoint errors are (or may be treated
// as) transport-domain errors. For example, an endpoint error will count
// against a circuit breaker error count. Therefore, it's almost certainly
// better to return service (business logic) errors in the response object. This
// means we have to do a bit more work in the HTTP response encoder to detect
// e.g. a not-found error and provide a proper HTTP status code. That work is
// done with the errorer interface, in transport.go.

func makePostProfileEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postProfileRequest)
		e := s.PostProfile(ctx, req.Profile)
		return postProfileResponse{Err: e}, nil
	}
}

type getProfileRequest struct {
	ID string
}

type getProfileResponse struct {
	Profile Profile `json:"profile,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r getProfileResponse) error() error { return r.Err }

func makeGetProfileEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getProfileRequest)
		p, e := s.GetProfile(ctx, req.ID)
		return getProfileResponse{Profile: p, Err: e}, nil
	}
}

type putProfileRequest struct {
	ID      string
	Profile Profile
}

type putProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r putProfileResponse) error() error { return nil }

func makePutProfileEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putProfileRequest)
		e := s.PutProfile(ctx, req.ID, req.Profile)
		return putProfileResponse{Err: e}, nil
	}
}

type patchProfileRequest struct {
	ID      string
	Profile Profile
}

type patchProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r patchProfileResponse) error() error { return r.Err }

func makePatchProfileEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(patchProfileRequest)
		e := s.PatchProfile(ctx, req.ID, req.Profile)
		return patchProfileResponse{Err: e}, nil
	}
}

type deleteProfileRequest struct {
	ID string
}

type deleteProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteProfileResponse) error() error { return r.Err }

func makeDeleteProfileEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteProfileRequest)
		e := s.DeleteProfile(ctx, req.ID)
		return deleteProfileResponse{Err: e}, nil
	}
}

type getAddressesRequest struct {
	ProfileID string
}

type getAddressesResponse struct {
	Addresses []Address `json:"addresses,omitempty"`
	Err       error     `json:"err,omitempty"`
}

func (r getAddressesResponse) error() error { return r.Err }

func makeGetAddressesEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getAddressesRequest)
		a, e := s.GetAddresses(ctx, req.ProfileID)
		return getAddressesResponse{Addresses: a, Err: e}, nil
	}
}

type getAddressRequest struct {
	ProfileID string
	AddressID string
}

type getAddressResponse struct {
	Address Address `json:"address,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r getAddressResponse) error() error { return r.Err }

func makeGetAddressEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getAddressRequest)
		a, e := s.GetAddress(ctx, req.ProfileID, req.AddressID)
		return getAddressResponse{Address: a, Err: e}, nil
	}
}

type postAddressRequest struct {
	ProfileID string
	Address   Address
}

type postAddressResponse struct {
	Err error `json:"err,omitempty"`
}

func (r postAddressResponse) error() error { return r.Err }

func makePostAddressEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postAddressRequest)
		e := s.PostAddress(ctx, req.ProfileID, req.Address)
		return postAddressResponse{Err: e}, nil
	}
}

type deleteAddressRequest struct {
	ProfileID string
	AddressID string
}

type deleteAddressResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteAddressResponse) error() error { return r.Err }

func makeDeleteAddressEndpoint(s ProfileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteAddressRequest)
		e := s.DeleteAddress(ctx, req.ProfileID, req.AddressID)
		return deleteAddressResponse{Err: e}, nil
	}
}
