package gateways_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gsaaraujo/pay-bank-api/internal/gateways"
	testhelpers "github.com/gsaaraujo/pay-bank-api/internal/test_helpers"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/stretchr/testify/suite"
)

type AwsSecretsGatewaySuite struct {
	suite.Suite
	secretsClient     *secretsmanager.Client
	awsSecretsGateway gateways.AwsSecretsGateway
}

func (a *AwsSecretsGatewaySuite) SetupSuite() {
	localstackContainer := testhelpers.NewLocalstackContainer()

	utils.ThrowOnError(os.Setenv("AWS_REGION", "us-east-1"))
	utils.ThrowOnError(os.Setenv("AWS_ACCESS_KEY_ID", "test"))
	utils.ThrowOnError(os.Setenv("AWS_SECRET_ACCESS_KEY", "test"))
	utils.ThrowOnError(os.Setenv("AWS_ENDPOINT_URL", localstackContainer.Url()))
	utils.ThrowOnError(os.Setenv("AWS_SECRET_MANAGER_NAME", "secret-us-east-1-local-app"))

	awsConfig := utils.GetOrThrow(config.LoadDefaultConfig(context.TODO()))
	a.secretsClient = secretsmanager.NewFromConfig(awsConfig)

	a.awsSecretsGateway = gateways.NewAwsSecretsGateway(a.secretsClient)
}

func (a *AwsSecretsGatewaySuite) SetupTest() {
	_, err := a.secretsClient.DeleteSecret(context.Background(), &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String("secret-us-east-1-local-app"),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	a.Require().NoError(err)
}

func (a *AwsSecretsGatewaySuite) Test1() {
	a.Run("given that the secrets were added, when getting, then returns secret", func() {
		utils.GetOrThrow(a.secretsClient.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
			Name: aws.String("secret-us-east-1-local-app"),
			SecretString: aws.String(`
				{
					"ANY_INT": 1,
					"ANY_STRING": "abc",
					"ANY_BOOLEAN": true
				}
			`),
		}))

		anyInt := a.awsSecretsGateway.Get("ANY_INT")
		a.Require().Equal(float64(1), anyInt)

		anyString := a.awsSecretsGateway.Get("ANY_STRING")
		a.Require().Equal("abc", anyString)

		anyBoolean := a.awsSecretsGateway.Get("ANY_BOOLEAN")
		a.Require().Equal(true, anyBoolean)
	})
}

func (a *AwsSecretsGatewaySuite) Test2() {
	a.Run("given that the secrets were not added, when getting, then panics", func() {
		utils.GetOrThrow(a.secretsClient.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
			Name: aws.String("secret-us-east-1-local-app"),
			SecretString: aws.String(`
				{
					"ANY_INT": 1,
					"ANY_STRING": "abc",
					"ANY_BOOLEAN": true
				}
			`),
		}))

		a.PanicsWithValue("POSTGRES_URL secret not found", func() { a.awsSecretsGateway.Get("POSTGRES_URL") })
	})
}

func TestAwsSecretsGateway(t *testing.T) {
	suite.Run(t, new(AwsSecretsGatewaySuite))
}
