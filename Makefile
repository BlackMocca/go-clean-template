app_name=app
port=3000
expose=3000
db_driver=postgres

dev.serve:
	docker-compose up $(app_name)

dev.down:
	docker-compose down 

prod.build:
	docker build ./ -t $(app_name)

prod.run:
	docker run --rm --name $(app_name) -p $(port):$(expose) -d $(app_name) 

prod.down:
	docker stop $(app_name)

prod.serve:
	make prod.build app_name=$(app_name)
	make prod.run app_name=$(app_name) port=$(port) expose=$(expose)

install-migration:
	docker exec -it $(app_name) sh -c "wget https://github.com/golang-migrate/migrate/releases/download/v4.6.2/migrate.linux-amd64.tar.gz"
	docker exec -it $(app_name) sh -c "tar xf migrate.linux-amd64.tar.gz"
	docker exec -it $(app_name) mv migrate.linux-amd64 /go/bin/migrate
	docker exec -it $(app_name) rm -f migrate.linux-amd64.tar.gz

app.migration.create:
	docker exec -it $(app_name) migrate create -ext $(db_driver) -dir database/migrations -seq create_$(table)_table

app.migration.up:
	docker exec -it $(app_name) migrate -database "$(db_url)" -path database/migrations up

app.migration.fix:
	docker exec -it $(app_name) migrate -database "$(db_url)" -path database/migrations force $(version)

app.migration.down:
	docker exec -it $(app_name) migrate -database "$(db_url)" -path database/migrations down
