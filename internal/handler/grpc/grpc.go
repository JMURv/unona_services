package grpc

import (
	"context"
	pb "github.com/JMURv/unona/services/api/pb"
	controller "github.com/JMURv/unona/services/internal/controller/rating"
	metrics "github.com/JMURv/unona/services/internal/metrics/prometheus"
	mdl "github.com/JMURv/unona/services/pkg/model"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"sync"
	"time"
)

type Connection struct {
	stream   pb.Broadcast_CreateStreamServer
	userUUID uuid.UUID
	active   bool
	error    chan error
}

type Pool struct {
	Connection []*Connection
}

type Handler struct {
	pb.BroadcastServer
	pb.NotificationsServer
	ctrl *controller.Controller
	pool *Pool
}

func New(ctrl *controller.Controller) *Handler {
	return &Handler{
		ctrl: ctrl,
		pool: &Pool{
			Connection: []*Connection{},
		},
	}
}

func (h *Handler) CreateStream(pbConn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream:   stream,
		userUUID: uuid.MustParse(pbConn.UserUuid),
		active:   true,
		error:    make(chan error),
	}

	h.pool.Connection = append(h.pool.Connection, conn)
	log.Printf("UserID: %v has been connected\n", pbConn.UserUuid)
	return <-conn.error
}

func (h *Handler) Broadcast(ctx context.Context, msg *mdl.Notification) error {
	var statusCode codes.Code
	start := time.Now()

	// TODO Request to users to get receiver username?
	span := opentracing.GlobalTracer().StartSpan("notifications.Broadcast.handler")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(start), int(statusCode), "Broadcast")
	}()

	var wg sync.WaitGroup
	for _, conn := range h.pool.Connection {
		wg.Add(1)
		go func(msg *mdl.Notification, conn *Connection) {
			defer wg.Done()
			if conn.active && (conn.userUUID == msg.ReceiverUUID || msg.ForBoth && conn.userUUID != msg.UserUUID) {
				log.Printf("Sending message to: %v from %v\n", conn.userUUID, msg.UserUUID)
				if err := conn.stream.Send(mdl.NotificationToProto(msg)); err != nil {
					log.Printf("Error with Stream: %v - Error: %v\n", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	wg.Wait()
	return nil
}

func (h *Handler) ListUserNotifications(ctx context.Context, req *pb.ByUserUUIDRequest) (*pb.ListNotificationResponse, error) {
	var statusCode codes.Code
	start := time.Now()

	span := opentracing.GlobalTracer().StartSpan("notifications.ListUserNotifications.handler")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(start), int(statusCode), "ListUserNotifications")
	}()

	userUUID := req.UserUuid
	if req == nil || userUUID == "" {
		statusCode = codes.InvalidArgument
		return nil, status.Errorf(statusCode, "nil req or empty id")
	}

	n, err := h.ctrl.ListUserNotifications(ctx, uuid.MustParse(userUUID))
	if err != nil {
		statusCode = codes.Internal
		span.SetTag("error", true)
		return nil, status.Errorf(statusCode, err.Error())
	}

	statusCode = codes.OK
	return &pb.ListNotificationResponse{Notifications: mdl.NotificationsToProto(*n)}, nil
}

func (h *Handler) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*pb.Empty, error) {
	var statusCode codes.Code
	start := time.Now()

	span := opentracing.GlobalTracer().StartSpan("notifications.DeleteNotification.handler")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(start), int(statusCode), "DeleteNotification")
	}()

	if req == nil || req.Id == 0 {
		statusCode = codes.InvalidArgument
		return nil, status.Errorf(statusCode, "nil req or empty id")
	}

	if err := h.ctrl.DeleteNotification(ctx, req.Id); err != nil {
		statusCode = codes.Internal
		span.SetTag("error", true)
		return nil, status.Errorf(statusCode, err.Error())
	}

	statusCode = codes.OK
	return &pb.Empty{}, nil
}

func (h *Handler) DeleteAllNotifications(ctx context.Context, req *pb.ByUserUUIDRequest) (*pb.Empty, error) {
	var statusCode codes.Code
	start := time.Now()

	span := opentracing.GlobalTracer().StartSpan("notifications.DeleteAllNotifications.handler")
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(start), int(statusCode), "DeleteAllNotifications")
	}()

	if req == nil || req.UserUuid == "" {
		statusCode = codes.InvalidArgument
		return nil, status.Errorf(statusCode, "nil req or empty id")
	}

	if err := h.ctrl.DeleteAllNotifications(ctx, uuid.MustParse(req.UserUuid)); err != nil {
		statusCode = codes.Internal
		span.SetTag("error", true)
		return nil, status.Errorf(statusCode, err.Error())
	}

	statusCode = codes.OK
	return &pb.Empty{}, nil
}
