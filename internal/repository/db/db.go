package db

import (
	"context"
	"fmt"
	repo "github.com/JMURv/unona/services/internal/repository"
	conf "github.com/JMURv/unona/services/pkg/config"
	"github.com/JMURv/unona/services/pkg/model"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

type Repository struct {
	conn *gorm.DB
}

func New(conf *conf.DBConfig) *Repository {
	DSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%v/%s",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Database,
	)

	conn, err := gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = conn.AutoMigrate(
		&model.Notification{},
	)
	if err != nil {
		log.Fatal(err)
	}

	return &Repository{conn: conn}
}

func (r *Repository) ListUserNotifications(ctx context.Context, userUUID uuid.UUID) (*[]*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.ListUserNotifications.repo")
	defer span.Finish()

	var n []*model.Notification
	if err := r.conn.Where("ReceiverID=?", userUUID).Find(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *Repository) CreateNotification(ctx context.Context, notify *model.Notification) (*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.CreateNotification.repo")
	defer span.Finish()

	if notify.Type == "" {
		return nil, repo.ErrTypeIsRequired
	}

	if notify.UserUUID == uuid.Nil {
		return nil, repo.ErrUserIDIsRequired
	}

	if notify.ReceiverUUID == uuid.Nil {
		return nil, repo.ErrIRecieverIDIsRequired
	}

	if notify.Message == "" {
		return nil, repo.ErrMessageIsRequired
	}

	notify.CreatedAt = time.Now()
	if err := r.conn.Create(notify).Error; err != nil {
		return nil, err
	}

	return notify, nil
}

func (r *Repository) DeleteNotification(ctx context.Context, notificationID uint64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.DeleteNotification.repo")
	defer span.Finish()

	if err := r.conn.Delete(&model.Notification{}, notificationID).Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteAllNotifications(ctx context.Context, userUUID uuid.UUID) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "notifications.DeleteAllNotifications.repo")
	defer span.Finish()

	if err := r.conn.Where("ReceiverID=?", userUUID).Delete(&model.Notification{}).Error; err != nil {
		return err
	}
	return nil
}
