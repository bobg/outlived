test:
	go build ./cmd/outlived
	cd web; npm run-script build
	./outlived -test serve

web:
	cd web; npm run-script build
	cp web/public/index.html web/build

check:
	go vet ./...
	cd web; npx tsc --noEmit

deploy:
	go build ./cmd/outlived
	cd web; npm run-script ship
	cp web/public/index.html web/build
	gcloud app deploy --project outlived-163105 app.yaml cron.yaml
