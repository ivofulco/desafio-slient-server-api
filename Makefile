
pre-req:
	sudo apt-get install build-essential -y

db-up:
	docker compose up -d;

server-up:
	cd ./server; echo "I'm in server folder"; \
	go mod tidy; \
	go run main.go CGO_ENABLED=1;

client-up:
	sleep 10s;
	cd ./client; echo "I'm in client folder"; \
	go run main.go CGO_ENABLED=1;
	echo "Arquivo gerado com sucesso na pasta client..."

.PHONY: desafio

desafio: db-up
	make -j 2 server-up client-up




# go mod init github.com/ivofulco/desafio-slient-server-api
# go mod tidy