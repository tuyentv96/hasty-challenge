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

I seprate the API and worker for some reason:
- API server will receive job request from user, save to database and send to queue.
- Worker will consume message from queue and execute job.
- We can develop and deploy the API server and worker independently.
- Modify and deploy API server will not require redeploy worker instance.
- API server and worker can scale independently

Jobs have four status:
`created`: When job was created
`running`: When worker claim and processing job
`success`: When job was success
`failed`: When job was failed or exceed timeout

System flow:
- When user send create request to server, server will response a `job_id`.
- Server save job to database and send to queue.
- Workers claim the job and mark the job `success` or `failed`

Design thining:
- Jobs with same `object_id` in time windows of 5 minutes will return the same `job_id`.
- When worker consume message from `Redis Queue`. It begin a transaction, claim the job for execute, set job `status` to `running`. And set job `status` to `success` or `failed` when done. So worker can rerun the job event when crash/restart.
- Job execution timeout will be set by env `JOB_TIMEOUT` in seconds.
- The prefetch limit `JOB_PREFETCH` is a limit number of jobs that a worker can reserve for itself

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
I use Redis for queue
I use Postgresql for persistent.

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

`start_time` is the time when job was claimed.
`end_time` is the time when job was done.
`message` will store error message when job was failed or job exceed timeout message.

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
Run tests will require docker for create Postgresql and Redis container.
```
go test -v ./...
```