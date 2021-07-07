test:
	go build ./cmd/outlived
	cd web; npm run-script build
	./outlived -test serve

web:
	cd web; npm run-script build

check:
	go vet ./...
	cd web; npx tsc --noEmit

deploy:
	go build ./cmd/outlived
	cd web; npm run-script ship
	gcloud app deploy --project outlived-163105 app.yaml cron.yaml
