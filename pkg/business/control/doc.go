//go:build tools
// +build tools

package control

import _ "github.com/golang/mock/mockgen/model"

//go:generate go run github.com/golang/mock/mockgen -destination mock/userrepo_mock.go -source=userrepo.go -package controlmock
