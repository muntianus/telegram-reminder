deploy:
	./deploy.sh

logs:
	docker logs --timestamps --tail 100 telegram-reminder

clean:
	docker compose down

.PHONY: deploy logs clean
