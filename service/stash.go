package service

import "net/http"

type stashService struct {
	client *http.Client
}
