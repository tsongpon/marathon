#!/usr/bin/env bash
set -euo pipefail

# --- Configuration ---
PROJECT_ID="abc-tech-477502"
REGION="asia-southeast1"
REPO_NAME="marathon"
SERVICE_NAME="marathon"
IMAGE_NAME="marathon"

IMAGE_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${IMAGE_NAME}:latest"

echo "==> Building container image..."
docker build --platform linux/amd64 -t "${IMAGE_URI}" .

echo "==> Pushing image to Artifact Registry..."
docker push "${IMAGE_URI}"

echo "==> Deploying to Cloud Run..."
gcloud run deploy "${SERVICE_NAME}" \
  --image "${IMAGE_URI}" \
  --region "${REGION}" \
  --project "${PROJECT_ID}" \
  --platform managed \
  --allow-unauthenticated \
  --port 1323 \
  --set-env-vars "GOOGLE_CLOUD_PROJECT=${PROJECT_ID},GOOGLE_CREDENTIALS_JSON=${GOOGLE_CREDENTIALS_JSON},ON_CALL_CALENDAR_ID=${ON_CALL_CALENDAR_ID},AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID},AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY},AWS_REGION=${AWS_REGION:-ap-southeast-1}"

echo "==> Done!"
