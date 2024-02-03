package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
)

type User struct {
	id   int
	name string
}

type ApiCall struct {
	id            int
	endpoint      string
	callTimestamp time.Time
}

type DBEnvVars struct {
	user     string
	password string
	dbname   string
	port     string
}

const DATE_FORMAT = "2006-01-02T15:04:05.99999999"

const STAGE_DEV = "dev"

func getCsvHeader() []string {
	return []string{"ID", "Endpoint", "Date"}
}

func connectToDB() (*sql.DB, error) {
	dbEnvVars := &DBEnvVars{
		user:     os.Getenv("POSTGRES_USER"),
		password: os.Getenv("POSTGRES_PASSWORD"),
		dbname:   os.Getenv("POSTGRES_DB"),
		port:     os.Getenv("POSTGRES_PORT"),
	}

	connectionString := fmt.Sprintf("host=db port=%s user=%s password=%s dbname=%s", dbEnvVars.port, dbEnvVars.user, dbEnvVars.password, dbEnvVars.dbname)

	if os.Getenv("STAGE") == STAGE_DEV {
		connectionString += " sslmode=disable"
	}

	return sql.Open("postgres", connectionString)
}

func main() {
	fmt.Println("Started user-history-module execution at ", time.Now().UTC().Format(DATE_FORMAT))
	db, err := connectToDB()

	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()
	db.SetMaxOpenConns(90) // Set maximum number of open connections
	db.SetMaxIdleConns(10) // Set maximum number of idle connections

	userRows, err := db.Query("SELECT * FROM \"user\"")
	if err != nil {
		fmt.Println("Error executing user query:", err)
		return
	}

	var cfg aws.Config
	if os.Getenv("STAGE") == STAGE_DEV {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: "http://minio:9000", HostnameImmutable: true}, nil
				})),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	}

	if err != nil {
		fmt.Println("Unable to load AWS configuration:", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)

	saveApiCallsFilesForAllUsers(userRows, db, s3Client)
}

func saveApiCallsFilesForAllUsers(userRows *sql.Rows, db *sql.DB, s3Client *s3.Client) {
	defer userRows.Close()
	wg := new(sync.WaitGroup)

	for userRows.Next() {
		var user User
		if err := userRows.Scan(&user.id, &user.name); err != nil {
			fmt.Println("Error scanning user row:", err)
			return
		}

		wg.Add(1)
		go saveApiCallsFileForUser(user, wg, db, s3Client)
	}

	wg.Wait()

	if err := userRows.Err(); err != nil {
		fmt.Println("Error iterating over user rows:", err)
		return
	}

	fmt.Println("Ended user-history-module execution at ", time.Now().UTC().Format(DATE_FORMAT))
}

func saveApiCallsFileForUser(user User, wg *sync.WaitGroup, db *sql.DB, s3Client *s3.Client) {
	defer wg.Done()
	filename := fmt.Sprintf("user_%d_%s.csv", user.id, user.name)
	bucketname := "user-history"

	rows, err := db.Query(fmt.Sprintf("SELECT id, endpoint, call_timestamp as callTimestamp FROM api_calls where user_id = %d", user.id))
	if err != nil {
		fmt.Printf("Error executing query api_calls for user %d: %v\n", user.id, err)
		return
	}

	defer rows.Close()

	var content bytes.Buffer
	csvwriter := csv.NewWriter(&content)

	if err := csvwriter.Write(getCsvHeader()); err != nil {
		fmt.Printf("Error writing header to file for user %d: %v\n", user.id, err)
		return
	}

	for rows.Next() {
		var apiCall ApiCall
		if err := rows.Scan(&apiCall.id, &apiCall.endpoint, &apiCall.callTimestamp); err != nil {
			fmt.Printf("Error scanning row for user %d: %v\n", user.id, err)
			return
		}

		row := []string{strconv.Itoa(apiCall.id), apiCall.endpoint, apiCall.callTimestamp.UTC().Format(DATE_FORMAT)}

		if err := csvwriter.Write(row); err != nil {
			fmt.Printf("Error writing record to file for user %d: %v\n", user.id, err)
			return
		}
	}
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketname,
		Key:    &filename,
		Body:   bytes.NewReader(content.Bytes()),
	})
	if err != nil {
		fmt.Printf("Unable to upload CSV buffer to S3 for user %d: %v\n", user.id, err)
		return
	}
}
