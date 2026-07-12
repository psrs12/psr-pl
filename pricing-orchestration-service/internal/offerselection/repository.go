package offerselection

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var ErrNotFound = errors.New("selected offer not found")

// Repository is create/get/update only, same no-delete pattern as
// application-management-service's Application repository — this record
// is part of an application's history, not something to discard.
type Repository interface {
	Create(ctx context.Context, offer *SelectedOffer) error
	Get(ctx context.Context, applicationID string) (*SelectedOffer, error)
	Update(ctx context.Context, offer *SelectedOffer) error
}

type dynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoRepository(client *dynamodb.Client, tableName string) *dynamoRepository {
	return &dynamoRepository{client: client, tableName: tableName}
}

func (r *dynamoRepository) Create(ctx context.Context, offer *SelectedOffer) error {
	item, err := attributevalue.MarshalMap(offer)
	if err != nil {
		return fmt.Errorf("marshalling selected offer: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: awsStringPtr("attribute_not_exists(applicationId)"),
	})
	if err != nil {
		return fmt.Errorf("creating selected offer: %w", err)
	}
	return nil
}

func (r *dynamoRepository) Get(ctx context.Context, applicationID string) (*SelectedOffer, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"applicationId": applicationID})
	if err != nil {
		return nil, fmt.Errorf("marshalling key: %w", err)
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving selected offer: %w", err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}

	var offer SelectedOffer
	if err := attributevalue.UnmarshalMap(out.Item, &offer); err != nil {
		return nil, fmt.Errorf("unmarshalling selected offer: %w", err)
	}
	return &offer, nil
}

func (r *dynamoRepository) Update(ctx context.Context, offer *SelectedOffer) error {
	item, err := attributevalue.MarshalMap(offer)
	if err != nil {
		return fmt.Errorf("marshalling selected offer: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: awsStringPtr("attribute_exists(applicationId)"),
	})
	if err != nil {
		return fmt.Errorf("updating selected offer: %w", err)
	}
	return nil
}

func awsStringPtr(s string) *string { return &s }
