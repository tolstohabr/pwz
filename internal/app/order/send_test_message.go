package order

import (
	"context"
	"log"

	desc "PWZ1.0/pkg/order"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func (i *Implementation) SendMessage(ctx context.Context, req *desc.MessageRequest) (*desc.MessageResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if version, ok := md["client-version"]; ok {
			log.Printf("Client version is %s", version[0])
		}
	}

	log.Printf("SendMessage: %v", req)
	return &desc.MessageResponse{Id: uuid.New().ID()}, nil //uuid.New().ID() тут просто генерируем рандомный id
}
