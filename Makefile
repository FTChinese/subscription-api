build_dir := build
config_file := api.toml

api_name := subscription-api
sandbox_name := subs_sandbox
consumer_name := iap-kafka-consumer

api_dev_out := $(build_dir)/$(api_name)
api_prod_out := $(build_dir)/linux/$(api_name)

sandbox_dev_out := $(build_dir)/$(sandbox_name)
sandbox_prod_out := $(build_dir)/linux$(sandbox_name)
sandbox_src_dir := ./cmd/subs_sandbox

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
linux_sandbox := GOOS=linux GOARCH=amd64 go build -o $(api_prod_out) $(ldflags) -tags production -v $(sandbox_src_dir)

go_version := go1.15

# Development
.PHONY: dev-api
dev-api :
	go build -o $(api_dev_out) $(ldflags) -v .

.PHONY: dev-sandbox
dev-sandbox :
	go build -o $(sandbox_dev_out) $(ldflasg) -v $(sandbox_src_dir)

.PHONY: dev-consumer
dev-consumer :
	go build -o $(consumer_dev_out) $(ldflags) -v $(consumer_src_dir)

# Run development build
.PHONY: run-api
run-api :
	./$(api_dev_out)

.PHONY: run-sandbox
run-sandbox :
	./$(sandbox_dev_out) -sandbox

.PHONY: run-consumer
run-consumer :
	./$(consumer_dev_out)

# Cross compiling linux on Mac for dev.
.PHONY: linux-api
linux-api :
	$(linux_api)

.PHONY: linux-sandbox
linux-sandbox :
	$(linux_sandbox)

.PHONY: linux-consumer
linux-consumer :
	$(linux_consumer)

# For CI/CD
.PHONY: install-go
install-go:
	gvm install $(go_version)
	gvm use $(go_version)

.PHONY: build-api
build-api : install-go
	$(linux_api)

.PHONY: build-sandbox
build-sandbox : install-go
	$(linux_sandbox)

.PHONY: build-consumer
build-consumer : install-go
	$(linux_consumer)

.PHONY: config
config :
	rsync -v tk11:/home/node/config/$(config_file) ./$(build_dir)
	rsync -v ./$(build_dir)/$(config_file) ucloud:/home/node/config

.PHONY: publish-api
publish-api :
	ssh ucloud "rm -f /home/node/go/bin/$(api_name).bak"
	rsync -v $(api_prod_out) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(api_name) ucloud:/home/node/go/bin/$(api_name).bak"

publish-sandbox :
	ssh ucloud "rm -f /home/node/go/bin/$(sandbox_name).bak"
	rsync -v $(sandbox_prod_out) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(sandbox_name) ucloud:/home/node/go/bin/$(sandbox_name).bak"

publish-consumer :
	ssh ucloud "rm -f /home/node/go/bin/$(consumer_name).bak"
	rsync -v $(consumer_prod_out) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(consumer_name) ucloud:/home/node/go/bin/$(consumer_name).bak"

.PHONY: restart-api
restart-api :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(api_name).bak $(api_name)"
	ssh ucloud supervisorctl restart $(api_name)

restart-sandbox :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(sandbox_name).bak $(sandbox_name)"
	ssh ucloud supervisorctl restart $(sandbox_name)

restart-consumer :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(consumer_name).bak $(sandbox_name)"
	ssh ucloud supervisorctl restart $(consumer_name)

.PHONY: clean
clean :
	go clean -x
	rm build/*
