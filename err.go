package hermes

import "errors"

var (
	ErrKeyNotFound          = errors.New("key not found")
	ErrKeyExpired           = errors.New("key expired")
	ErrKeyExists            = errors.New("key already exists")
	ErrInvalidType          = errors.New("invalid data type")
	ErrValueMismatch        = errors.New("value mismatch")
	ErrInvalidValueType     = errors.New("invalid value type")
	ErrContextCanceled      = errors.New("operation canceled")
	ErrInvalidTTL           = errors.New("invalid TTL value")
	ErrEmptyList            = errors.New("list is empty")
	ErrEmptyValues          = errors.New("empty value")
	ErrInvalidKey           = errors.New("invalid key")
	ErrTransactionNotActive = errors.New("transaction is not active")
	ErrTransactionFailed    = errors.New("transaction failed")
)

func IsKeyNotFound(err error) bool {
	return errors.Is(err, ErrKeyNotFound)
}

func IsKeyExpired(err error) bool {
	return errors.Is(err, ErrKeyExpired)
}

func IsKeyExists(err error) bool {
	return errors.Is(err, ErrKeyExists)
}

func IsInvalidType(err error) bool {
	return errors.Is(err, ErrInvalidType)
}

func IsValueMismatch(err error) bool {
	return errors.Is(err, ErrValueMismatch)
}

func IsInvalidValueType(err error) bool {
	return errors.Is(err, ErrInvalidValueType)
}

func IsContextCanceled(err error) bool {
	return errors.Is(err, ErrContextCanceled)
}

func IsInvalidTTL(err error) bool {
	return errors.Is(err, ErrInvalidTTL)
}

func IsEmptyList(err error) bool {
	return errors.Is(err, ErrEmptyList)
}

func IsInvalidKey(err error) bool {
	return errors.Is(err, ErrInvalidKey)
}

func IsTransactionNotActive(err error) bool {
	return errors.Is(err, ErrTransactionNotActive)
}

func IsTransactionFailed(err error) bool {
	return errors.Is(err, ErrTransactionFailed)
}
