# Hasty challenge

Job service

## 1. Requirements
* Golang 1.16.x
* Postgres (Database)
* Redis (Queue)
* Docker
* Docker Compose

## 2. Run

Run docker compose
```
docker-compose up --scale job-worker=4
```