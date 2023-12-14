TEST_DIR?="./..."

lint:
	golangci-lint run --fix --fast

test:
	go test -race -coverprofile=coverage.out -covermode=atomic $(TEST_DIR)

bench:
	go test -v -bench=. -benchmem $(TEST_DIR)

run_example:
	cd examples/${TEST_DIR} && go mod tidy && go run -ldflags="-X 'github.com/go-thor/thor/build.ID=1234567' -X 'github.com/go-thor/thor/build.Name=demo_name' -X 'github.com/go-thor/thor/build.Version=1.0.0' -X 'github.com/go-thor/thor/build.Namespace=demo_namespace'" .

pre_commit:
	pre-commit run --all-files

pre_commit_install:
	pre-commit install

pre_commit_uninstall:
	pre-commit uninstall
