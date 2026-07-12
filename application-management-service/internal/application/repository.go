package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var ErrNotFound = errors.New("application not found")

// Repository intentionally exposes no delete operation: submitted
// applications must be retained for recordkeeping/adverse-action
// compliance and are never removed once persisted.
type Repository interface {
	Create(ctx context.Context, app *Application) error
	Update(ctx context.Context, app *Application) error
	Get(ctx context.Context, id string) (*Application, error)
}

type dynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoRepository(client *dynamodb.Client, tableName string) *dynamoRepository {
	return &dynamoRepository{client: client, tableName: tableName}
}

func (r *dynamoRepository) Create(ctx context.Context, app *Application) error {
	item, err := attributevalue.MarshalMap(app)
	if err != nil {
		return fmt.Errorf("marshalling application: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: awsStringPtr("attribute_not_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}
	return nil
}

func (r *dynamoRepository) Update(ctx context.Context, app *Application) error {
	item, err := attributevalue.MarshalMap(app)
	if err != nil {
		return fmt.Errorf("marshalling application: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: awsStringPtr("attribute_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return nil
}

func (r *dynamoRepository) Get(ctx context.Context, id string) (*Application, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"id": id})
	if err != nil {
		return nil, fmt.Errorf("marshalling key: %w", err)
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving application: %w", err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}

	var app Application
	if err := attributevalue.UnmarshalMap(out.Item, &app); err != nil {
		return nil, fmt.Errorf("unmarshalling application: %w", err)
	}
	return &app, nil
}

func awsStringPtr(s string) *string { return &s }
