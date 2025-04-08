package errors

import "errors"

var KeyNotExists = errors.New("key not exists")
var UserNotExists = errors.New("user not exists")
var WrongPassword = errors.New("wrong password")
var UserAlreadyExists = errors.New("user already exists")
var InvalidAmount = errors.New("invalid amount")
var InvalidCurrency = errors.New("invalid currency")
var InsufficientFunds = errors.New("insufficient funds")
