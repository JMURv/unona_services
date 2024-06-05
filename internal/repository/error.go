package repository

import "errors"

var ErrNotFound = errors.New("user not found")
var ErrTypeIsRequired = errors.New("type is required")
var ErrUserIDIsRequired = errors.New("userID is required")
var ErrIRecieverIDIsRequired = errors.New("receiverID is required")
var ErrMessageIsRequired = errors.New("message is required")
