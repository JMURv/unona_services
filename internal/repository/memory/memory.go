package memory

import (
	"context"
	repo "github.com/JMURv/unona/services/internal/repository"
	"github.com/JMURv/unona/services/pkg/model"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"sync"
	"time"
)

type Repository struct {
	sync.RWMutex
	data map[uint64]*model.Notification
}

func New() *Repository {
	return &Repository{data: map[uint64]*model.Notification{}}
}

func (r *Repository) ListUserNotifications(ctx context.Context, userUUID uuid.UUID) (*[]*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.ListUserNotifications.repo")
	defer span.Finish()

	r.RLock()
	defer r.RUnlock()
	n := make([]*model.Notification, 0, len(r.data))
	for _, v := range r.data {
		if v.ReceiverUUID == userUUID {
			n = append(n, v)
		}
	}
	return &n, nil
}

func (r *Repository) CreateNotification(ctx context.Context, notify *model.Notification) (*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.CreateNotification.repo")
	defer span.Finish()

	notify.ID = uint64(time.Now().Unix())

	if notify.Type == "" {
		return nil, repo.ErrTypeIsRequired
	}

	if notify.UserUUID == uuid.Nil {
		return nil, repo.ErrUserIDIsRequired
	}

	if notify.ReceiverUUID == uuid.Nil {
		return nil, repo.ErrIReceiverIDIsRequired
	}

	if notify.Message == "" {
		return nil, repo.ErrMessageIsRequired
	}

	notify.CreatedAt = time.Now()

	r.Lock()
	defer r.Unlock()
	r.data[notify.ID] = notify

	return notify, nil
}

func (r *Repository) DeleteNotification(ctx context.Context, notificationID uint64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.DeleteNotification.repo")
	defer span.Finish()

	r.Lock()
	defer r.Unlock()
	delete(r.data, notificationID)
	return nil
}

func (r *Repository) DeleteAllNotifications(ctx context.Context, userUUID uuid.UUID) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.DeleteAllNotifications.repo")
	defer span.Finish()

	r.Lock()
	defer r.Unlock()
	for i, n := range r.data {
		if userUUID == n.ReceiverUUID {
			delete(r.data, i)
			return nil
		}
	}
	return repo.ErrNotFound
}
