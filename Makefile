
bin/viamwpasupplicantmgr: *.go cmd/module/*.go
	go build -o bin/viamwpasupplicantmgr cmd/module/cmd.go

test:
	go test

lint:
	gofmt -w -s .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

module: bin/viamwpasupplicantmgr
	tar czf module.tar.gz bin/viamwpasupplicantmgr
