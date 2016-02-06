package main

import (
	"errors"
	"sync"

	"golang.org/x/net/context"
)

// ProfileService is a simple CRUD interface for user profiles.
type ProfileService interface {
	PostProfile(ctx context.Context, p Profile) error
	GetProfile(ctx context.Context, id string) (Profile, error)
	PutProfile(ctx context.Context, id string, p Profile) error
	PatchProfile(ctx context.Context, id string, p Profile) error
	DeleteProfile(ctx context.Context, id string) error
	GetAddresses(ctx context.Context, profileID string) ([]Address, error)
	GetAddress(ctx context.Context, profileID string, addressID string) (Address, error)
	PostAddress(ctx context.Context, profileID string, a Address) error
	DeleteAddress(ctx context.Context, profileID string, addressID string) error
}

// Profile represents a single user profile.
// ID should be globally unique.
type Profile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Addresses []Address `json:"addresses,omitempty"`
}

// Address is a field of a user profile.
// ID should be unique within the profile (at a minimum).
type Address struct {
	ID       string `json:"id"`
	Location string `json:"location,omitempty"`
}

var (
	errInconsistentIDs = errors.New("inconsistent IDs")
	errAlreadyExists   = errors.New("already exists")
	errNotFound        = errors.New("not found")
)

type inmemService struct {
	mtx sync.RWMutex
	m   map[string]Profile
}

func newInmemService() ProfileService {
	return &inmemService{
		m: map[string]Profile{},
	}
}

func (s *inmemService) PostProfile(ctx context.Context, p Profile) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[p.ID]; ok {
		return errAlreadyExists
	}
	s.m[p.ID] = p
	return nil
}

func (s *inmemService) GetProfile(ctx context.Context, id string) (Profile, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	p, ok := s.m[id]
	if !ok {
		return Profile{}, errNotFound
	}
	return p, nil
}

func (s *inmemService) PutProfile(ctx context.Context, id string, p Profile) error {
	if id != p.ID {
		return errInconsistentIDs
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[id]; ok {
		return errAlreadyExists
	}
	s.m[id] = p
	return nil
}

func (s *inmemService) PatchProfile(ctx context.Context, id string, p Profile) error {
	return s.PutProfile(ctx, id, p) // perhaps more granular behavior is needed here
}

func (s *inmemService) DeleteProfile(ctx context.Context, id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[id]; !ok {
		return errNotFound
	}
	delete(s.m, id)
	return nil
}

func (s *inmemService) GetAddresses(ctx context.Context, profileID string) ([]Address, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	p, ok := s.m[profileID]
	if !ok {
		return []Address{}, errNotFound
	}
	return p.Addresses, nil
}

func (s *inmemService) GetAddress(ctx context.Context, profileID string, addressID string) (Address, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	p, ok := s.m[profileID]
	if !ok {
		return Address{}, errNotFound
	}
	for _, address := range p.Addresses {
		if address.ID == addressID {
			return address, nil
		}
	}
	return Address{}, errNotFound
}

func (s *inmemService) PostAddress(ctx context.Context, profileID string, a Address) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	p, ok := s.m[profileID]
	if !ok {
		return errNotFound
	}
	for _, address := range p.Addresses {
		if address.ID == a.ID {
			return errAlreadyExists
		}
	}
	p.Addresses = append(p.Addresses, a)
	s.m[profileID] = p
	return nil
}

func (s *inmemService) DeleteAddress(ctx context.Context, profileID string, addressID string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	p, ok := s.m[profileID]
	if !ok {
		return errNotFound
	}
	newAddresses := make([]Address, 0, len(p.Addresses))
	for _, address := range p.Addresses {
		if address.ID == addressID {
			continue // delete
		}
		newAddresses = append(newAddresses, address)
	}
	if len(newAddresses) == len(p.Addresses) {
		return errNotFound
	}
	p.Addresses = newAddresses
	s.m[profileID] = p
	return nil
}
