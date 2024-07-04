build:
	cd events && go get -u && go mod tidy && CGO_ENABLED=0 go build -o com.utmstack.events.plugin
	cd alerts && go get -u && go mod tidy && go build -o com.utmstack.alerts.plugin
	cd inputs && go get -u && go mod tidy && go build -o com.utmstack.inputs.plugin
	cd geolocation && go get -u && go mod tidy && go build -o com.utmstack.geolocation.plugin