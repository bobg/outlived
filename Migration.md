# Outlived App Engine to Firebase Migration Guide

This document describes the steps required to migrate the Outlived application from Google App Engine (GAE) to Firebase Hosting and Google Cloud Run.

## Prerequisites

Ensure you have the following tools installed and authenticated:
- [Google Cloud CLI (`gcloud`)](https://cloud.google.com/sdk/docs/install)
- [Firebase CLI (`firebase`)](https://firebase.google.com/docs/cli)
- Docker (for building the backend image)

Ensure you are logged in:
```bash
gcloud auth login
gcloud auth configure-docker us-central1-docker.pkg.dev
firebase login
```

Set the default Google Cloud project:
```bash
gcloud config set project outlived-163105
```

---

## Step 1: Enable Google Cloud APIs

Enable the necessary APIs for Artifact Registry, Cloud Run, Cloud Tasks, and Cloud Scheduler:

```bash
gcloud services enable \
  artifactregistry.googleapis.com \
  run.googleapis.com \
  cloudtasks.googleapis.com \
  cloudscheduler.googleapis.com
```

---

## Step 2: Build and Deploy the Backend Container to Cloud Run

### 1. Create Artifact Registry Repository
Create a Docker repository in Artifact Registry to store your backend images:

```bash
gcloud artifacts repositories create outlived \
  --repository-format=docker \
  --location=us-central1 \
  --description="Outlived Docker Repository"
```

### 2. Build and Push the Docker Image
Build the container image using the project's [Dockerfile](file:///Users/bobg/go/src/github.com/bobg/outlived/Dockerfile):

```bash
docker build -t us-central1-docker.pkg.dev/outlived-163105/outlived/backend:latest .
docker push us-central1-docker.pkg.dev/outlived-163105/outlived/backend:latest
```

### 3. Deploy to Cloud Run
Deploy the backend image to Cloud Run. Make sure to specify the service name `outlived-backend` (as referenced in [firebase.json](file:///Users/bobg/go/src/github.com/bobg/outlived/firebase.json)):

```bash
gcloud run deploy outlived-backend \
  --image=us-central1-docker.pkg.dev/outlived-163105/outlived/backend:latest \
  --platform=managed \
  --region=us-central1 \
  --allow-unauthenticated
```

*Note: The backend service url will be mapped automatically by Firebase Hosting, so keeping the service unauthenticated is acceptable since API/Cron endpoints are protected by the `X-Outlived-Key` header check.*

---

## Step 3: Deploy the Frontend to Firebase Hosting

### 1. Compile the React Frontend
Build the frontend package using `esbuild` to generate the production bundle (`web/public/bundle.js`):

```bash
cd web
npm install
npm run-script ship
cd ..
```

### 2. Deploy to Firebase
Initialize the project with Firebase Hosting (if not already done), and deploy it:

```bash
firebase use outlived-163105
firebase deploy --only hosting
```

---

## Step 4: Configure Cloud Scheduler Jobs (Cron Replacement)

Since App Engine is no longer running, `cron.yaml` needs to be replaced with Cloud Scheduler jobs. 

Get the `master-key` value from your Datastore (you can use `outlived admin get master-key`). Set it as a variable for the commands below:

```bash
MASTER_KEY="your-master-key-here"
```

Create the scheduler jobs targeting the Firebase Hosting production domain (`https://outlived.net`) with the `X-Outlived-Key` header:

### 1. Scrape Launcher
Runs monthly on the 5th and 20th at 01:00:
```bash
gcloud scheduler jobs create http scrape-launcher \
  --schedule="0 1 5,20 * *" \
  --uri="https://outlived.net/t/scrape" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1
```

### 2. Figure Expirer
Runs daily at midnight:
```bash
gcloud scheduler jobs create http figure-expirer \
  --schedule="0 0 * * *" \
  --uri="https://outlived.net/t/expire" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1
```

### 3. Timezone Mailing Jobs
Create a Cloud Scheduler job for each timezone defined in [cron.yaml](file:///Users/bobg/go/src/github.com/bobg/outlived/cron.yaml) to trigger daily at 07:00 in that specific timezone:

```bash
# Pacific/Honolulu
gcloud scheduler jobs create http mailing-honolulu \
  --schedule="0 7 * * *" \
  --time-zone="Pacific/Honolulu" \
  --uri="https://outlived.net/t/send?tzoffset=-36000&tzname=Pacific/Honolulu" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# America/Los_Angeles
gcloud scheduler jobs create http mailing-los-angeles \
  --schedule="0 7 * * *" \
  --time-zone="America/Los_Angeles" \
  --uri="https://outlived.net/t/send?tzoffset=-28800&tzname=America/Los_Angeles" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# America/Chicago
gcloud scheduler jobs create http mailing-chicago \
  --schedule="0 7 * * *" \
  --time-zone="America/Chicago" \
  --uri="https://outlived.net/t/send?tzoffset=-21600&tzname=America/Chicago" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# America/Puerto_Rico
gcloud scheduler jobs create http mailing-puerto-rico \
  --schedule="0 7 * * *" \
  --time-zone="America/Puerto_Rico" \
  --uri="https://outlived.net/t/send?tzoffset=-14400&tzname=America/Puerto_Rico" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Atlantic/South_Georgia
gcloud scheduler jobs create http mailing-south-georgia \
  --schedule="0 7 * * *" \
  --time-zone="Atlantic/South_Georgia" \
  --uri="https://outlived.net/t/send?tzoffset=-7200&tzname=Atlantic/South_Georgia" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# UTC
gcloud scheduler jobs create http mailing-utc \
  --schedule="0 7 * * *" \
  --time-zone="UTC" \
  --uri="https://outlived.net/t/send?tzoffset=0&tzname=UTC" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Africa/Cairo
gcloud scheduler jobs create http mailing-cairo \
  --schedule="0 7 * * *" \
  --time-zone="Africa/Cairo" \
  --uri="https://outlived.net/t/send?tzoffset=7200&tzname=Africa/Cairo" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Europe/Samara
gcloud scheduler jobs create http mailing-samara \
  --schedule="0 7 * * *" \
  --time-zone="Europe/Samara" \
  --uri="https://outlived.net/t/send?tzoffset=14400&tzname=Europe/Samara" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Asia/Dhaka
gcloud scheduler jobs create http mailing-dhaka \
  --schedule="0 7 * * *" \
  --time-zone="Asia/Dhaka" \
  --uri="https://outlived.net/t/send?tzoffset=21600&tzname=Asia/Dhaka" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Asia/Hong_Kong
gcloud scheduler jobs create http mailing-hong-kong \
  --schedule="0 7 * * *" \
  --time-zone="Asia/Hong_Kong" \
  --uri="https://outlived.net/t/send?tzoffset=28800&tzname=Asia/Hong_Kong" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Australia/Sydney
gcloud scheduler jobs create http mailing-sydney \
  --schedule="0 7 * * *" \
  --time-zone="Australia/Sydney" \
  --uri="https://outlived.net/t/send?tzoffset=36000&tzname=Australia/Sydney" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Pacific/Fiji
gcloud scheduler jobs create http mailing-fiji \
  --schedule="0 7 * * *" \
  --time-zone="Pacific/Fiji" \
  --uri="https://outlived.net/t/send?tzoffset=43200&tzname=Pacific/Fiji" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1

# Pacific/Kiritimati
gcloud scheduler jobs create http mailing-kiritimati \
  --schedule="0 7 * * *" \
  --time-zone="Pacific/Kiritimati" \
  --uri="https://outlived.net/t/send?tzoffset=50400&tzname=Pacific/Kiritimati" \
  --http-method="GET" \
  --headers="X-Outlived-Key=${MASTER_KEY}" \
  --location=us-central1
```

---

## Step 5: Clean Up App Engine (Optional)

Once migration is verified and traffic has fully cut over to Firebase/Cloud Run, you can disable or delete the App Engine application to prevent double-billing. Note that Google Cloud Datastore (now Firestore in Datastore mode) is independent of the App Engine application status and will continue to work.
