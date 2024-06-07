package model

import (
	pb "github.com/JMURv/unona/services/api/pb"
	"github.com/google/uuid"
)

func NotificationsToProto(n []*Notification) []*pb.Notification {
	res := make([]*pb.Notification, 0, len(n))
	for _, v := range n {
		res = append(res, NotificationToProto(v))
	}
	return res
}
func NotificationToProto(n *Notification) *pb.Notification {
	return &pb.Notification{
		Id:           n.ID,
		Type:         n.Type,
		UserUuid:     n.UserID.String(),
		ReceiverUuid: n.ReceiverID.String(),
		Message:      n.Message,
	}
}

func NotificationFromProto(n *pb.Notification) *Notification {
	return &Notification{
		ID:         n.Id,
		Type:       n.Type,
		UserID:     uuid.MustParse(n.UserUuid),
		ReceiverID: uuid.MustParse(n.ReceiverUuid),
		Message:    n.Message,
	}
}
