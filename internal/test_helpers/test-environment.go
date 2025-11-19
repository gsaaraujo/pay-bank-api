package testhelpers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gsaaraujo/pay-bank-api/internal"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type TestEnvironment struct {
	baseUrl                string
	client                 *http.Client
	awsConfig              aws.Config
	pgxPool                *pgxpool.Pool
	redisClient            *redis.Client
	rabbitmqConn           *amqp091.Connection
	postgresContainerUrl   string
	localstackContainerUrl string
	wiremockContainerUrl   string
	redisContainerUrl      string
	rabbitmqContainerUrl   string
}

func NewTestEnvironment() *TestEnvironment {
	return &TestEnvironment{}
}

func (t *TestEnvironment) Start() {
	t.postgresContainerUrl = NewPostgresContainer().url
	t.localstackContainerUrl = NewLocalstackContainer().url
	t.wiremockContainerUrl = NewWiremockContainer().url
	t.redisContainerUrl = NewRedisContainer().url
	t.rabbitmqContainerUrl = NewRabbitmqContainer().url

	utils.ThrowOnError(os.Setenv("AWS_REGION", "us-east-1"))
	utils.ThrowOnError(os.Setenv("AWS_ACCESS_KEY_ID", "test"))
	utils.ThrowOnError(os.Setenv("AWS_SECRET_ACCESS_KEY", "test"))
	utils.ThrowOnError(os.Setenv("AWS_ENDPOINT_URL", t.localstackContainerUrl))
	utils.ThrowOnError(os.Setenv("AWS_SECRET_MANAGER_NAME", "secret-us-east-1-local-app"))
	utils.ThrowOnError(os.Setenv("TERN_MIGRATIONS_PATH", "../migrations"))
	utils.ThrowOnError(os.Setenv("ZIPCODE_URL", t.wiremockContainerUrl))

	t.awsConfig = utils.GetOrThrow(config.LoadDefaultConfig(context.TODO()))
	t.createSecrets()
	utils.ThrowOnError(t.runMigrations())

	t.pgxPool = utils.GetOrThrow(pgxpool.New(context.Background(), t.postgresContainerUrl))
	t.rabbitmqConn = utils.GetOrThrow(amqp091.Dial(t.rabbitmqContainerUrl))

	redisParsedUrl := utils.GetOrThrow(redis.ParseURL(t.redisContainerUrl))
	t.redisClient = redis.NewClient(&redis.Options{
		Addr:     redisParsedUrl.Addr,
		Password: redisParsedUrl.Password,
	})

	t.startHttpServer()
}

func (t *TestEnvironment) createSecrets() {
	secretsClient := secretsmanager.NewFromConfig(t.awsConfig)

	utils.GetOrThrow(secretsClient.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
		Name: aws.String("secret-us-east-1-local-app"),
		SecretString: aws.String(fmt.Sprintf(`
			{
				"REDIS_URL": "%s",
				"POSTGRES_URL": "%s",
				"RABBITMQ_URL": "%s",
				"ACCESS_TOKEN_SIGNING_KEY": "81c4a8d5b2554de4ba736e93255ba633"
			}
		`, t.redisContainerUrl, t.postgresContainerUrl, t.rabbitmqContainerUrl)),
	}))
}

func (t *TestEnvironment) runMigrations() error {
	urlParsed := utils.GetOrThrow(url.Parse(t.postgresContainerUrl))

	utils.ThrowOnError(os.Setenv("PGUSER", "postgres"))
	utils.ThrowOnError(os.Setenv("PGPASSWORD", "postgres"))
	utils.ThrowOnError(os.Setenv("PGHOST", urlParsed.Hostname()))
	utils.ThrowOnError(os.Setenv("PGPORT", urlParsed.Port()))
	utils.ThrowOnError(os.Setenv("PGDATABASE", "postgres"))

	migrations := ""
	if _, ok := os.LookupEnv("TERN_MIGRATIONS_PATH"); !ok {
		return errors.New("TERN_MIGRATIONS_PATH environment variable not found")
	}

	migrations = os.Getenv("TERN_MIGRATIONS_PATH")
	cmd := exec.Command("tern", "migrate", "-m", migrations)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %s, output: %s", err.Error(), string(output))
	}

	return nil
}

func (t *TestEnvironment) startHttpServer() {
	httpServer := internal.NewHttpServer()
	httpServer.Ready()
	server := httptest.NewServer(httpServer.Echo())

	t.baseUrl = server.URL
	t.client = server.Client()
}

func (s *TestEnvironment) BaseUrl() string {
	return s.baseUrl
}

func (s *TestEnvironment) WiremockContainerUrl() string {
	return s.wiremockContainerUrl
}

func (s *TestEnvironment) Client() *http.Client {
	return s.client
}

func (s *TestEnvironment) PgxPool() *pgxpool.Pool {
	return s.pgxPool
}

func (s *TestEnvironment) RedisClient() *redis.Client {
	return s.redisClient
}

func (s *TestEnvironment) RabbitmqConn() *amqp091.Connection {
	return s.rabbitmqConn
}
