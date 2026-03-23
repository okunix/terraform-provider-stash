package service

import "context"

type StashService interface {
	GetStashByID(ctx context.Context, stashID string) error
	GetStashSecretEntry(ctx context.Context, stashID, entryName string) error
	CreateStash(ctx context.Context) error
	DeleteStash(ctx context.Context, stashID string) error
	AddMember(ctx context.Context, stashID, userID string) error
	RemoveMember(ctx context.Context, stashID, userID string) error
}

type UserService interface {
	GetUserByUsername(ctx context.Context, username string)
	GetToken(ctx context.Context, username, password string) (string, error)
}
