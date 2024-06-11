package model

import (
	pb "github.com/JMURv/unona/services/api/pb"
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
		UserUuid:     n.UserUUID.String(),
		ReceiverUuid: n.ReceiverUUID.String(),
		Message:      n.Message,
		ForBoth:      n.ForBoth,
	}
}
