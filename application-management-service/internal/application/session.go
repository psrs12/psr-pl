package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var ErrSessionInvalid = errors.New("session token is invalid or expired")

// SessionStore is a small, consumer-owned interface for the applicant
// self-service login flow: exchange verified identity for a short-lived
// token, and validate that token on subsequent status checks.
type SessionStore interface {
	Create(ctx context.Context, applicationID string, ttl time.Duration) (token string, err error)
	Validate(ctx context.Context, token string) (applicationID string, err error)
}

type sessionRecord struct {
	ID            string `dynamodbav:"id"`
	ApplicationID string `dynamodbav:"applicationId"`
	ExpiresAt     int64  `dynamodbav:"ttl"`
}

type dynamoSessionStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoSessionStore stores sessions in the same table as Application
// records, distinguished by an "id" prefix, consistent with single-table
// design. The table's DynamoDB TTL attribute must be configured on "ttl"
// for expired sessions to be reaped automatically.
func NewDynamoSessionStore(client *dynamodb.Client, tableName string) *dynamoSessionStore {
	return &dynamoSessionStore{client: client, tableName: tableName}
}

func (s *dynamoSessionStore) Create(ctx context.Context, applicationID string, ttl time.Duration) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}

	record := sessionRecord{
		ID:            sessionItemID(token),
		ApplicationID: applicationID,
		ExpiresAt:     time.Now().Add(ttl).Unix(),
	}
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return "", fmt.Errorf("marshalling session: %w", err)
	}

	if _, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	}); err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}
	return token, nil
}

func (s *dynamoSessionStore) Validate(ctx context.Context, token string) (string, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"id": sessionItemID(token)})
	if err != nil {
		return "", fmt.Errorf("marshalling session key: %w", err)
	}

	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key:       key,
	})
	if err != nil {
		return "", fmt.Errorf("retrieving session: %w", err)
	}
	if out.Item == nil {
		return "", ErrSessionInvalid
	}

	var record sessionRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return "", fmt.Errorf("unmarshalling session: %w", err)
	}
	if time.Now().Unix() >= record.ExpiresAt {
		return "", ErrSessionInvalid
	}
	return record.ApplicationID, nil
}

func sessionItemID(token string) string {
	return "SESSION#" + token
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
