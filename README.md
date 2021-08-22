# Job service

## 1. Requirements
* Golang 1.16.x
* Postgres (Database)
* Redis (Queue)
* Docker
* Docker Compose

## 2. Quick start
Run with default:
```
docker-compose up
```

Run with scale worker:
```
docker-compose up --scale job-worker=4
```

## 3. Architecture

I separate the API and worker for some reason:
- API server will receive job request from the user, save to database and send to queue.
- Worker will consume messages from the queue and execute the job.
- We can develop and deploy the API server and worker independently.
- Modify and deploy API server will not require redeploy worker instance.
- API server and worker can scale independently

Jobs have four statuses:
- `created`: When the job was created
- `running`: When workers claim and processing job
- `success`: When the job was success
- `failed`: When the job was failed or exceed the timeout

System flow:
- When the user sends create request to the server, the server will respond a `job_id`.
- Server saves the job to the database and sends it to the queue.
- Workers claim the job and mark the job `success` or `failed`

Design thinking:
- I use Redis for queue and Postgresql for persistent.
- Jobs with the same `object_id` in time windows of 5 minutes will return the same `job_id`.
- When workers consume messages from `Redis Queue`. It begins a transaction, claims the job for execution, sets job `status` to `running`. And set job `status` to `success` or `failed` when done. So the worker can rerun the job event when crash/restart.
- Job execution timeout will be set by env `JOB_TIMEOUT` in seconds.
- The env prefetch limit `JOB_PREFETCH` is a limited number of jobs that a worker can reserve for itself

## 4. API desgin:
Create Job API
```
curl --location --request POST 'localhost:3000/v1/jobs' \
--header 'Content-Type: application/json' \
--data-raw '{
    "object_id": 1
}'
```

Get Job API
```
curl --location --request GET 'localhost:3000/v1/jobs/1'
```

## 5. Database:
Database schema:
```
CREATE TABLE IF NOT EXISTS "jobs" (
"id" serial PRIMARY KEY,
"object_id" integer NOT NULL,
"status" text NOT NULL,
"start_time" timestamp(6),
"end_time" timestamp(6),
"message" TEXT,
"created_at" timestamp(6) NOT NULL DEFAULT timezone('utc'::text, now())
);
```

`start_time` is the time when the job was claimed.
`end_time` is the time when the job was done.
`message` will store an error message when the job was failed or the job exceeds the timeout message.

Code Structure:
```
project
└─── cmd   // commands and DI
└─── db     // migrations
└─── config // configuration
└─── jobs   // main logic
└─── utils
│   README.md
│   Dockerfile   
│   docker-compose.yml
│   main.go
```


## 6. Run tests:
Run tests will require docker to create Postgresql and Redis containers.
```
go test -v ./...
```
