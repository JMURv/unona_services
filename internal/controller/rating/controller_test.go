package user

import (
	broker "github.com/JMURv/unona/media/internal/broker/memory"
	cache "github.com/JMURv/unona/media/internal/cache/memory"
	repo "github.com/JMURv/unona/media/internal/repository/memory"
)

var svc *Controller

func init() {
	r := repo.New()
	c := cache.New()
	b := broker.New()
	svc = New(r, c, b)
}
