build:
	cd events && go get -u && go mod tidy && go build -o com.utmstack.events.plugin
	cd alerts && go get -u && go mod tidy && go build -o com.utmstack.alerts.plugin
	cd inputs && go get -u && go mod tidy && go build -o com.utmstack.inputs.plugin
	cd geolocation && go get -u && go mod tidy && go build -o com.utmstack.geolocation.plugin
	cd stats && go get -u && go mod tidy && go build -o com.utmstack.stats.plugin
	cd gcp && go get -u && go mod tidy && go build -o com.utmstack.gcp.plugin