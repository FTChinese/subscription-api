build_dir := build
BINARY := subscription-api

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD} -X main.lastCommit=${COMMIT}"

doc_file := subscription_api
inputfiles := frontmatter.md README.md

.PHONY: build linux deploy attack report lastcommit mkbuild clean
build :
	go build $(LDFLAGS) -o $(build_dir)/$(BINARY) -v .

run :
	./$(build_dir)/${BINARY}

deploy : linux
	rsync -v $(build_dir)/linux/$(BINARY) nodeserver:/home/node/go/bin/

# Copy env varaible to server
config :
	rsync -v ../.env nodeserver:/home/node/go

linux : 
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(build_dir)/linux/$(BINARY) -v .

pdf : mkbuild
	pandoc -s --toc --pdf-engine=xelatex -o $(build_dir)/$(doc_file).pdf $(inputfiles)

mkbuild :
	mkdir -p $(build_dir)

clean :
	go clean -x
	rm build/*