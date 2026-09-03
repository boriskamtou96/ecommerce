#!/bin/bash

# Create bucket
awslocal s3 mb s3://ecommerce-uploads

echo "LocalStack initialization complete"

# Create SQS queue
awslocal sqs create-queue --queue-name ecommerce-queue