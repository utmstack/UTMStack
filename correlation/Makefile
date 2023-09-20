reinstall: uninstall install

install:
	GOOS=linux go build .
	docker build -t utmstack.azurecr.io/correlation:testing .
	docker stack deploy utmstack --compose-file utmstack.yml

uninstall:
	docker stack rm utmstack
	sleep 15s
