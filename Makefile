test:
	go build ./cmd/outlived
	(cd web; npm run-script build)
	./outlived serve -test

check:
	go vet ./...
	(cd web; npx tsc --project . --noEmit)
