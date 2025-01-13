package storage

import "errors"

var ErrNotUniqueID = errors.New("a record with this ID already exists")
var ErrStorageLimitExceeded = errors.New("storage limit exceeded")
