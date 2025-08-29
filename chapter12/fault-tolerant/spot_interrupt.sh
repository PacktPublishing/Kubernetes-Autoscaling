#!/bin/bash

# AWS FIS Spot Interruption Script
INSTANCE_NAME_TAG="karpenter-default"  # Change this to match your instance's Name tag
REGION="eu-west-1"                     # Change to your region

export AWS_PAGER=""

echo "AWS FIS Spot Instance Interruption Script"
echo "=============================================="

echo "Creating FIS IAM role..."

ROLE_NAME="FIS-SpotInterruption-Role"
TRUST_POLICY='{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "fis.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}'

# Create role if it doesn't exist
if ! aws iam get-role --role-name $ROLE_NAME &> /dev/null; then
    aws iam create-role \
        --role-name $ROLE_NAME \
        --assume-role-policy-document "$TRUST_POLICY" \
        --region $REGION 

    aws iam attach-role-policy \
        --role-name $ROLE_NAME \
        --policy-arn "arn:aws:iam::aws:policy/service-role/AWSFaultInjectionSimulatorEC2Access" \
        --region $REGION
    
    echo "IAM role created"
    sleep 10
else
    echo "IAM role already exists"
fi

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ROLE_ARN="arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"

echo "Creating FIS experiment template..."

TEMPLATE_JSON='{
    "description": "Spot Instance Interruption Test",
    "targets": {
        "SpotInstances": {
            "resourceType": "aws:ec2:spot-instance",
            "resourceTags": {
                "Name": "'$INSTANCE_NAME_TAG'"
            },
            "filters": [
                {
                    "path": "State.Name",
                    "values": ["running"]
                }
            ],
            "selectionMode": "COUNT(1)"
        }
    },
    "actions": {
        "interrupt": {
            "actionId": "aws:ec2:send-spot-instance-interruptions",
            "description": "Interrupt one random spot instance",
            "parameters": {
                "durationBeforeInterruption": "PT2M"
            },
            "targets": {
                "SpotInstances": "SpotInstances"
            }
        }
    },
    "stopConditions": [
        {
            "source": "none"
        }
    ],
    "roleArn": "'$ROLE_ARN'"
}'

TEMPLATE_ID=$(aws fis create-experiment-template \
    --cli-input-json "$TEMPLATE_JSON" \
    --region $REGION \
    --query 'experimentTemplate.id' \
    --output text)

echo "Experiment template created: $TEMPLATE_ID"

echo "Starting spot interruption experiment..."
echo "Target: instances with Name tag '$INSTANCE_NAME_TAG'"

EXPERIMENT_ID=$(aws fis start-experiment \
    --experiment-template-id $TEMPLATE_ID \
    --region $REGION \
    --query 'experiment.id' \
    --output text)

echo "Experiment started: $EXPERIMENT_ID"
echo "Spot interruption initiated for instances tagged with Name: $INSTANCE_NAME_TAG"

(sleep 300 && aws fis delete-experiment-template --id $TEMPLATE_ID --region $REGION &> /dev/null) &

echo "Done. Monitor your application logs for graceful shutdown handling."
