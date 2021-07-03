test:
	go build ./cmd/outlived
	(cd web; npm run-script build)
	./outlived -test serve

check:
	go vet ./...
	(cd web; npx tsc --noEmit)

deploy:
	(cd web; yarn build)
	gcloud app deploy --project outlived-163105 app.yaml cron.yaml
