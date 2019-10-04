app_name=app
port=3000
expose=3000

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