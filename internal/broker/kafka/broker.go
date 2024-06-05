package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	ctrl "github.com/JMURv/unona/services/internal/controller/rating"
	hdlr "github.com/JMURv/unona/services/internal/handler/grpc"
	cfg "github.com/JMURv/unona/services/pkg/config"
	mdl "github.com/JMURv/unona/services/pkg/model"
	"log"
)

type Broker struct {
	topic    string
	consumer sarama.Consumer
	pc       map[string]sarama.PartitionConsumer

	ctrl    *ctrl.Controller
	handler *hdlr.Handler
}

func New(conf *cfg.KafkaConfig, ctrl *ctrl.Controller, handler *hdlr.Handler) *Broker {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(conf.Addrs, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}

	return &Broker{
		topic:    conf.NotificationTopic,
		consumer: consumer,
		ctrl:     ctrl,
		handler:  handler,
		pc:       make(map[string]sarama.PartitionConsumer),
	}
}

func (b *Broker) Start() {
	ctx := context.Background()
	topic := b.topic

	pc, err := b.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error consuming Kafka topic %s: %v", topic, err)
	}
	b.pc[topic] = pc

	log.Println("Kafka consumer started")
	for msg := range pc.Messages() {
		log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))
		switch msg.Topic {
		case topic:
			var n mdl.Notification
			if err := json.Unmarshal(msg.Value, &n); err != nil {
				log.Printf("Error unmarshalling notification: %v", err)
				continue
			}

			notification, err := b.ctrl.CreateNotification(ctx, &n)
			if err != nil {
				log.Printf("Error creating notification: %v", err)
			}

			err = b.handler.Broadcast(ctx, notification)
			if err != nil {
				log.Printf("Error broadcasting notification: %v", err)
			}
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
