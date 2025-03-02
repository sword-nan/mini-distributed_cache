package common

import (
	"errors"
)

var (
	//
	ErrTimeout = errors.New("operation timeout")
	//
	ErrPositiveParamNegative  = errors.New("param be a positive number but get a negative number")
	ErrKeyNotInDB             = errors.New("key not in db")
	ErrKeyNotInCache          = errors.New("key not in cache")
	ErrCacheCapacityNotEnough = errors.New("new entry is bigger than cache capacity")
	//
	ErrServiceNotExisted = errors.New("service is not existed")
	//
	ErrPeerRegistered    = errors.New("peer was already registered")
	ErrPeerNotRegistered = errors.New("peer is never registered")
	ErrNoPeerRegistered  = errors.New("no peer was registered")
)
