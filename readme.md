# User History Module

The User History Module is a dockerized Go application designed to interact with a PostgreSQL database to obtain all the API call data for each user, then generate and upload CSV files containing that data to an Amazon S3 bucket. The module uses Go's concurrency features to efficiently process user data and API call records.

## Concurrency vs Parallelism

During the development of this module, I explored both concurrency and parallelism to optimize the processing of users' API calls. After thorough testing with various database sizes and container configurations (adjusting CPU and memory allocations), I concluded that concurrency was the most effective approach.
In each configuration, concurrency consistently outperformed parallelism by at least 5 seconds.

**However...**
It's important to note that these tests were conducted in a local environment. In a real-world scenario, where database and S3 connection/upload times are expected to be higher, the performance difference between concurrency and parallelism may vary.

## Prerequisites

Before running the User History Module, ensure that the following prerequisites are met:

1. Docker is installed in your system.
2. Docker Compose is installed in your system.

## Run in Docker

To run the User History Module in Docker, follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/alejosagues/user-history-module.git
   ```

2. Navigate to the project directory:

   ```bash
   cd user-history-module
   ```

3. Build Docker Compose file:

   ```bash
   docker compose up --build
   ```

   Or, if you want to stop the application when it ends:

   ```bash
   docker compose up --build --exit-code-from app
   ```

4. Check the files on minio: In your browser, go to http://localhost:9001/, log in with the credentials provided in the `docker-compose.yml` file (found under `MINIO_ACCESS_KEY` and `MINIO_SECRET_KEY`) and enter the bucket "user-history".

5. Stop and remove the containers created when you finish:

   ```bash
   docker compose down
   ```

The module will connect to the PostgreSQL database, retrieve user data, generate CSV files containing API call records for each user, and upload them to the specified Amazon S3 bucket.

## Configuration

You can configure the User History Module by adjusting the following parameters:

    - Connection to PostgreSQL database: Set environment variables POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, and POSTGRES_PORT.
    - AWS: Set environment variable AWS_REGION for the desired AWS region and AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY for credentials.
    - Development mode: Set environment variable STAGE to dev for development mode. If it is in development mode, it will login to the PostgreSQL database using `sslmode=disable` and upload the CSV files to the minio url.

Additionally, you can modify the maximum number of open and idle connections to the PostgreSQL database by adjusting the values in the connectToDB function.
