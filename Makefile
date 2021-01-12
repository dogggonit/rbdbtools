all: bin/rbdbgen bin/rbdbdump

bin/rbdbgen: | bin requirements
	go build -o bin/rbdbgen cmd/rbdbgen/main.go

bin/rbdbdump: | bin requirements
	go build -o bin/rbdbdump cmd/rbdbdump/main.go

bin:
	mkdir $@

requirements:
	go mod download

clean:
	rm -rf bin