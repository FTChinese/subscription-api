build_dir := build
config_file := api.toml
api_name := subscription-api
consumer_name := iap-kafka-consumer

api_dev_out := $(build_dir)/$(api_name)
api_prod_out := $(build_dir)/linux/$(api_name)

consumer_dev_out := $(build_dir)/$(consumer_name)
consumer_prod_out := $(build_dir)/linux/$(consumer_name)
consumer_src_dir := ./cmd/iap-kafka-consumer/

LOCAL_CONFIG_FILE := $(HOME)/config/$(config_file)

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD} -X main.commit=${COMMIT}"

linux_api := GOOS=linux GOARCH=amd64 go build -o $(api_prod_out) $(ldflags) -tags production -v .
linux_consumer := GOOS=linux GOARCH=amd64 go build -o $(consumer_prod_out) $(ldflags) -tags production -v $(consumer_src_dir)

go_version := go1.15

.PHONY: local run linux config deploy build downconfig upconfig publish restart clean
# Development
dev-api :
	go build -o $(api_dev_out) $(ldflags) -v .

dev-consumer :
	go build -o $(consumer_dev_out) $(ldflags) -v $(consumer_src_dir)

# Run development build
run-api :
	./$(api_dev_out) -sandbox

run-consumer :
	./$(consumer_dev_out)

# Cross compiling linux on for dev.
linux-api :
	$(linux_api)

linux-consumer :
	$(linux_consumer)

#deploy :
#	rsync -v $(api_prod_out) tk11:/home/node/go/bin/
#	ssh tk11 supervisorctl restart $(api_name)

install-go:
	gvm install $(go_version)
	gvm use $(go_version)

# For CI/CD
build-api : install-go
	$(linux_api)

build-consumer : install-go
	$(linux_consumer)

config :
	rsync -v tk11:/home/node/config/$(config_file) ./$(build_dir)
	rsync -v ./$(build_dir)/$(config_file) ucloud:/home/node/config

publish-api :
	ssh ucloud "rm -f /home/node/go/bin/$(api_name).bak"
	rsync -v $(api_prod_out) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(api_name) ucloud:/home/node/go/bin/$(api_name).bak"
#	scp -rp $(LINUX_OUT) ucloud:/home/node/go/bin/$(BINARY).bak

restart-api :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(api_name).bak $(api_name)"
	ssh ucloud supervisorctl restart $(api_name)

clean :
	go clean -x
	rm build/*
