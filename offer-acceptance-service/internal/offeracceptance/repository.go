package offeracceptance

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrNotFound = errors.New("esign record not found")
var ErrAlreadySigned = errors.New("application already e-signed")

// Repository is create/get only. An e-signature is immutable once made —
// no update, no delete.
type Repository interface {
	Create(ctx context.Context, record *EsignRecord) error
	Get(ctx context.Context, applicationID string) (*EsignRecord, error)
}

type dynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoRepository(client *dynamodb.Client, tableName string) *dynamoRepository {
	return &dynamoRepository{client: client, tableName: tableName}
}

func (r *dynamoRepository) Create(ctx context.Context, record *EsignRecord) error {
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("marshalling esign record: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: awsStringPtr("attribute_not_exists(applicationId)"),
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return ErrAlreadySigned
		}
		return fmt.Errorf("creating esign record: %w", err)
	}
	return nil
}

func (r *dynamoRepository) Get(ctx context.Context, applicationID string) (*EsignRecord, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"applicationId": applicationID})
	if err != nil {
		return nil, fmt.Errorf("marshalling key: %w", err)
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving esign record: %w", err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}

	var record EsignRecord
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return nil, fmt.Errorf("unmarshalling esign record: %w", err)
	}
	return &record, nil
}

func awsStringPtr(s string) *string { return &s }
