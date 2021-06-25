stop:
	docker-compose stop
start: start-db
	docker-compose up konga kong
down:
	docker-compose down

start-db:
	docker-compose up -d pg
stop-db:
	docker-compose stop pg

exec-kong:
	@docker-compose exec kong /bin/bash
reload-kong:
	@docker-compose exec kong reload

start-konga:
	docker-compose up konga
exec-konga:
	docker-compose exec konga /bin/bash

start-cli:
	docker-compose up -d cli
stop-cli:
	docker-compose stop cli 
exec-cli: start-cli
	@docker-compose exec cli /bin/bash

migration: start-cli
	@docker-compose exec cli /bin/bash -c "kong migrations bootstrap" 
