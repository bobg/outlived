test:
	go build ./cmd/outlived
	(cd web; npm run-script build)
	./outlived serve -test

check:
	go vet ./...
	(cd web; npx tsc --project . --noEmit)

deploy:
	gcloud app deploy --project outlived-163105 app.yaml cron.yaml
