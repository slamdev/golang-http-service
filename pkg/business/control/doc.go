package control

import _ "github.com/golang/mock/mockgen/model"

//go:generate go run github.com/golang/mock/mockgen -destination userrepo_mock.go -source=userrepo.go -package control
