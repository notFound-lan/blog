start:
	@docker-compose up nginx
exec: 
	@docker-compose exec nginx /bin/bash
reload:
	@docker-compose exec nginx /bin/bash -c "nginx -s reload"
down:
	@docker-compose down
clean:
	docker-compose down --rmi all