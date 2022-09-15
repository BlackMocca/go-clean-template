service=go-clean-template
db_driver=postgres
path=migrations/database/postgres
version=latest
migrate_url=https://github.com/BlackMocca/migrate/releases/download/v5.0/migrate.linux-amd64

install-migration:
	docker exec -it $(service) sh -c "wget $(migrate_url)"
	docker exec -it $(service) sh -c "chmod +x migrate.linux-amd64"
	docker exec -it $(service) mv migrate.linux-amd64 /usr/local/bin/migrate

app.migration.create:
	migrate create -ext $(db_driver) -dir $(path) -seq create_$(table)_table

app.migration.up:
	migrate -database "$(db_url)"  -path $(path) up

app.migration.fix:
	migrate -database "$(db_url)"  -path $(path) force $(version)

app.migration.down:
	migrate -database "$(db_url)"  -path $(path) down

app.migration.seed:
	migrate -database "$(db_url)"  -path $(path) seed-up

mockery:
	mockery --all --dir service/$(service) --output service/$(service)/mocks

test:
	go test ./... -coverprofile cover.out
	go tool cover -html=cover.out -o coverage.html
	export unit_total=$$(go test ./... -v  | grep -c RUN) && echo "Unit Test Total: $$unit_total" && export coverage_total=$$(go tool cover -func cover.out | grep total | awk '{print $$3}') && echo "Coverage Total: $$coverage_total"

genproto:
	protoc -I . --go_out=plugins=grpc:proto/ proto/*.proto

