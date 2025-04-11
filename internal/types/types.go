package types

import "time"

type DataType int

const (
	String DataType = iota
	List
	Hash
	Set
)

type Entry struct {
	Value      interface{}
	Type       DataType
	Expiration time.Time
}
