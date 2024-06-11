package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	ctrl "github.com/JMURv/unona/services/internal/controller/rating"
	hdlr "github.com/JMURv/unona/services/internal/handler/grpc"
	"github.com/JMURv/unona/services/internal/smtp"
	cfg "github.com/JMURv/unona/services/pkg/config"
	mdl "github.com/JMURv/unona/services/pkg/model"
	"github.com/google/uuid"
	"log"
)

type EmailInterface interface {
	SendVerificationEmail(_ context.Context, userUUID uuid.UUID, photo []byte) error
	SendLoginEmail(_ context.Context, code uint64, toEmail string) error
	SendActivationCodeEmail(_ context.Context, code uint64, toEmail string) error
}

type Broker struct {
	notificationTopic        string
	verificationEmailTopic   string
	loginEmailTopic          string
	activationCodeEmailTopic string
	forgotPasswordEmailTopic string

	consumer sarama.Consumer
	pc       map[string]sarama.PartitionConsumer

	ctrl    *ctrl.Controller
	handler *hdlr.Handler
	email   *smtp.EmailServer
}

type DigitCode struct {
	Code    uint64 `json:"code"`
	ToEmail string `json:"to_email"`
}

func New(conf *cfg.KafkaConfig, ctrl *ctrl.Controller, handler *hdlr.Handler, email *smtp.EmailServer) *Broker {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(conf.Addrs, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}

	return &Broker{
		notificationTopic:        conf.NotificationTopic,
		verificationEmailTopic:   conf.VerificationEmailTopic,
		loginEmailTopic:          conf.LoginEmailTopic,
		activationCodeEmailTopic: conf.ActivationCodeEmailTopic,
		forgotPasswordEmailTopic: conf.ForgotPasswordEmailTopic,

		consumer: consumer,
		pc:       make(map[string]sarama.PartitionConsumer),
		ctrl:     ctrl,
		handler:  handler,
		email:    email,
	}
}

func (b *Broker) Start() {
	ctx := context.Background()
	for _, topic := range []string{b.notificationTopic, b.verificationEmailTopic, b.loginEmailTopic, b.activationCodeEmailTopic, b.forgotPasswordEmailTopic} {
		pc, err := b.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("Error consuming Kafka topic %s: %v", topic, err)
		}
		b.pc[topic] = pc

		switch topic {
		case b.notificationTopic:
			go b.handleNotifications(ctx, pc)
		case b.verificationEmailTopic:
			go b.handleVerificationEmail(ctx, pc)
		case b.loginEmailTopic:
			go b.handleLoginEmail(ctx, pc)
		case b.activationCodeEmailTopic:
			go b.handleActivationCodeEmail(ctx, pc)
		case b.forgotPasswordEmailTopic:
			go b.handleForgotPasswordEmail(ctx, pc)
		}
	}
}

func (b *Broker) Close() {
	for _, pc := range b.pc {
		if err := pc.Close(); err != nil {
			log.Printf("Error closing Kafka partition consumer: %v", err)
		}
	}
	if err := b.consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}
}

func (b *Broker) handleNotifications(ctx context.Context, pc sarama.PartitionConsumer) {
	log.Println("Notifications consumer started")

	for msg := range pc.Messages() {
		log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

		n := &mdl.Notification{}
		if err := json.Unmarshal(msg.Value, n); err != nil {
			log.Printf("Error unmarshalling notification: %v", err)
			continue
		}

		notification, err := b.ctrl.CreateNotification(ctx, n)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
		}

		if err = b.handler.Broadcast(ctx, notification); err != nil {
			log.Printf("Error broadcasting notification: %v", err)
		}
	}
}

func (b *Broker) handleVerificationEmail(ctx context.Context, pc sarama.PartitionConsumer) {
	log.Println("VerificationEmail consumer started")

	for msg := range pc.Messages() {
		log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

		e := &struct {
			UserUUID uuid.UUID `json:"user_uuid"`
			Photo    []byte    `json:"photo"`
		}{}
		if err := json.Unmarshal(msg.Value, e); err != nil {
			log.Printf("Error unmarshalling notification: %v", err)
			continue
		}

		err := b.email.SendVerificationEmail(ctx, e.UserUUID, e.Photo)
		if err != nil {
			log.Printf("Error sending verification email: %v", err)
		}
	}
}

func (b *Broker) handleLoginEmail(ctx context.Context, pc sarama.PartitionConsumer) {
	log.Println("LoginEmail consumer started")

	for msg := range pc.Messages() {
		log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

		e := &DigitCode{}
		if err := json.Unmarshal(msg.Value, e); err != nil {
			log.Printf("Error unmarshalling notification: %v", err)
			continue
		}

		err := b.email.SendLoginEmail(ctx, e.Code, e.ToEmail)
		if err != nil {
			log.Printf("Error sending verification email: %v", err)
		}
	}
}

func (b *Broker) handleActivationCodeEmail(ctx context.Context, pc sarama.PartitionConsumer) {
	log.Println("ActivationCodeEmail consumer started")

	for msg := range pc.Messages() {
		log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

		e := &DigitCode{}
		if err := json.Unmarshal(msg.Value, e); err != nil {
			log.Printf("Error unmarshalling notification: %v", err)
			continue
		}

		err := b.email.SendActivationCodeEmail(ctx, e.Code, e.ToEmail)
		if err != nil {
			log.Printf("Error sending activation code email: %v", err)
		}
	}
}

func (b *Broker) handleForgotPasswordEmail(ctx context.Context, pc sarama.PartitionConsumer) {
	log.Println("ForgotPasswordEmail consumer started")

	for msg := range pc.Messages() {
		log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

		e := &struct {
			Token   string `json:"token"`
			UID64   string `json:"uidb64"`
			ToEmail string `json:"to_email"`
		}{}
		if err := json.Unmarshal(msg.Value, e); err != nil {
			log.Printf("Error unmarshalling notification: %v", err)
			continue
		}

		err := b.email.SendForgotPasswordEmail(ctx, e.ToEmail, e.UID64, e.ToEmail)
		if err != nil {
			log.Printf("Error sending activation code email: %v", err)
		}
	}
}
