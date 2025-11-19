package gateways

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
)

type AwsSecretsGateway struct {
	secretsClient *secretsmanager.Client
}

func NewAwsSecretsGateway(secretsClient *secretsmanager.Client) AwsSecretsGateway {
	return AwsSecretsGateway{
		secretsClient: secretsClient,
	}
}

func (a *AwsSecretsGateway) Get(key string) any {
	if _, ok := os.LookupEnv("AWS_SECRET_MANAGER_NAME"); !ok {
		panic("AWS_SECRET_MANAGER_NAME environment variable not found")
	}

	secretValue := utils.GetOrThrow(a.secretsClient.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(os.Getenv("AWS_SECRET_MANAGER_NAME")),
	}))

	var secret map[string]any
	utils.ThrowOnError(json.Unmarshal([]byte(*secretValue.SecretString), &secret))

	value, exists := secret[key]
	if !exists {
		panic(fmt.Sprintf("%s secret not found", key))
	}

	return value
}
