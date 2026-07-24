package document

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var ErrNotFound = errors.New("document record not found")

// Repository is get/save only — documents arrive incrementally, so Save
// is an upsert (unlike offer-acceptance-service's create-only Repository,
// there's no immutability constraint here: the record legitimately
// changes as each document is submitted).
type Repository interface {
	Get(ctx context.Context, applicationID string) (*Record, error)
	Save(ctx context.Context, record *Record) error
}

type dynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoRepository(client *dynamodb.Client, tableName string) *dynamoRepository {
	return &dynamoRepository{client: client, tableName: tableName}
}

func (r *dynamoRepository) Get(ctx context.Context, applicationID string) (*Record, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"applicationId": applicationID})
	if err != nil {
		return nil, fmt.Errorf("marshalling key: %w", err)
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving document record: %w", err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}

	var record Record
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, fmt.Errorf("unmarshalling document record: %w", err)
	}
	return &record, nil
}

func (r *dynamoRepository) Save(ctx context.Context, record *Record) error {
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("marshalling document record: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("saving document record: %w", err)
	}
	return nil
}
