package user

import (
	"context"
	"fmt"
	"github.com/JMURv/unona/services/pkg/model"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"time"
)

const cacheKey = "notification:%v"

type CacheRepository interface {
	Get(ctx context.Context, key string) (*model.Notification, error)
	Set(ctx context.Context, t time.Duration, key string, r *model.Notification) error
	Delete(ctx context.Context, key string) error
}

type notificationsRepository interface {
	ListUserNotifications(ctx context.Context, userUUID uuid.UUID) (*[]*model.Notification, error)
	CreateNotification(ctx context.Context, data *model.Notification) (*model.Notification, error)
	DeleteNotification(ctx context.Context, notificationID uint64) error
	DeleteAllNotifications(ctx context.Context, userUUID uuid.UUID) error
}

type Controller struct {
	repo  notificationsRepository
	cache CacheRepository
}

func New(repo notificationsRepository, cache CacheRepository) *Controller {
	return &Controller{
		repo:  repo,
		cache: cache,
	}
}

func (c *Controller) ListUserNotifications(ctx context.Context, userUUID uuid.UUID) (*[]*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.ListUserNotifications.ctrl")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	res, err := c.repo.ListUserNotifications(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Controller) CreateNotification(ctx context.Context, data *model.Notification) (*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.CreateNotification.ctrl")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	res, err := c.repo.CreateNotification(ctx, data)
	if err != nil {
		return nil, err
	}

	err = c.cache.Set(ctx, time.Hour, fmt.Sprintf(cacheKey, res.ID), res)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (c *Controller) DeleteNotification(ctx context.Context, notificationID uint64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.DeleteNotification.ctrl")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	err := c.repo.DeleteNotification(ctx, notificationID)
	if err != nil {
		return err
	}

	if err = c.cache.Delete(ctx, fmt.Sprintf(cacheKey, notificationID)); err != nil {
		return err
	}

	return nil
}

func (c *Controller) DeleteAllNotifications(ctx context.Context, userUUID uuid.UUID) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.DeleteAllNotifications.ctrl")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	err := c.repo.DeleteAllNotifications(ctx, userUUID)
	if err != nil {
		return err
	}

	return nil
}
