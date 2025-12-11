#!/bin/bash

set -e

PROJECT_ID=${GCP_PROJECT_ID:-"your-project-id"}
REGION=${GCP_REGION:-"us-central1"}
SERVICE_NAME="api-backend"

echo "Deploying to Google Cloud Platform..."
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo ""

read -p "Continue with deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

echo "Building and pushing container..."
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME

echo "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars ENVIRONMENT=production

echo "Deployment complete!"
echo "Note: Don't forget to set up your DATABASE_URL environment variable in Cloud Run"
