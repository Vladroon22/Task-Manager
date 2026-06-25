.PHONY:

DB_URL = "postgres://postgres:55555@localhost:5432/postgres?sslmode=disable"

run:
	go run cmd/main.go

mig-up:
	migrate -database $(DB_URL) -path migrations up

mig-down:
	migrate -database $(DB_URL) -path migrations down

mig-force:
	migrate -database $(DB_URL) -path migrations force 1

swagger:
	swag init -g cmd/main.go -o docs --parseDependency --parseInternal

clean:
	rm -rf docs/

compose-run:
	sudo docker-compose up --build -d

compose-stop:
	sudo docker-compose down -v