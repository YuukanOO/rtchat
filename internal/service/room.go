// Package service just exposes everything needed to manage our tiny domain.
package service

import (
	"sync"

	"github.com/yuukanoo/rtchat/internal/crypto"
)

type (
	// Service exposed to create or get a room. It could be splitted but since
	// this application is lightweight, it should suffice for now.
	Service interface {
		// CreateRoom creates a room and returns its unique identity.
		CreateRoom() string
		// DeleteRoom deletes a room given its unique identity.
		DeleteRoom(string)
		// GetRoom retrieves a Room object from its identity.
		GetRoom(string) *Room
	}

	// Room object which contains TURN credential for this particular room.
	Room struct {
		ID         string
		Credential string
	}

	// service implements the Service interface with an in memory map, it should
	// suffice for now.
	service struct {
		mutex sync.RWMutex
		rooms map[string]*Room
	}
)

// New instantiates a new service to manage rooms.
func New() Service {
	return &service{
		rooms: make(map[string]*Room),
	}
}

func (s *service) CreateRoom() string {
	id := crypto.GenerateUID(32)

	r := &Room{
		ID:         id,
		Credential: crypto.GenerateUID(32), // And use a random string has the credential
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.rooms[id] = r

	return r.ID
}

func (s *service) DeleteRoom(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.rooms[id] = nil
	delete(s.rooms, id)
}

func (s *service) GetRoom(id string) *Room {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	room := s.rooms[id]
	return room
}
