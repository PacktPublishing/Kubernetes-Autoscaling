#!/bin/bash

kubectl get events --field-selector involvedObject.kind=NodeClaim -A --sort-by=.lastTimestamp -w