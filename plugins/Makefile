update: 
	cd events && go get -u && go mod tidy 
	cd alerts && go get -u && go mod tidy 
	cd inputs && go get -u && go mod tidy 
	cd geolocation && go get -u && go mod tidy 
	cd stats && go get -u && go mod tidy 
	cd gcp && go get -u && go mod tidy 
	cd aws && go get -u && go mod tidy 
	cd azure && go get -u && go mod tidy 
	cd sophos && go get -u && go mod tidy 
	cd o365 && go get -u && go mod tidy
	cd config && go get -u && go mod tidy
	cd bitdefender && go get -u && go mod tidy

build:
	cd events && go build -o com.utmstack.events.plugin
	cd alerts && go build -o com.utmstack.alerts.plugin
	cd inputs && go build -o com.utmstack.inputs.plugin
	cd geolocation && go build -o com.utmstack.geolocation.plugin
	cd stats && go build -o com.utmstack.stats.plugin
	cd gcp && CGO_ENABLED=0 go build -o com.utmstack.gcp.plugin
	cd aws && go build -o com.utmstack.aws.plugin
	cd azure && go build -o com.utmstack.azure.plugin
	cd sophos && go build -o com.utmstack.sophos.plugin
	cd o365 && go build -o com.utmstack.o365.plugin
	cd config && go build -o com.utmstack.config.plugin
	cd bitdefender && go build -o com.utmstack.config.plugin

update-build: update build
