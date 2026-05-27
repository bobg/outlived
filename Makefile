.PHONY: test web check deploy deploy-backend deploy-frontend liveupdates

test:
	go build ./cmd/outlived
	cd web; npm run-script build
	./outlived -test serve

web:
	cd web; npm run-script build

check:
	go vet ./...
	cd web; npx tsc --noEmit

deploy: deploy-backend deploy-frontend

deploy-backend:
	docker build -t us-central1-docker.pkg.dev/outlived-163105/outlived/backend:latest .
	docker push us-central1-docker.pkg.dev/outlived-163105/outlived/backend:latest
	gcloud run deploy outlived-backend \
		--image=us-central1-docker.pkg.dev/outlived-163105/outlived/backend:latest \
		--platform=managed \
		--region=us-central1 \
		--allow-unauthenticated

deploy-frontend:
	cd web; npm run-script ship
	firebase use outlived-163105
	firebase deploy --only hosting

liveupdates:
	inotifywait -e close_write -r web/src -m | (while read -r x; do echo $x; (cd web; npm run-script build); done)
