# Karpenter Events Reference

## Complete Events Table

| Event Reason | Event Type | Object Type | Description | Message | 
|--------------|------------|-------------|-------------|---------|
| DisruptionBlocked | Normal | Node, NodeClaim, NodePool | Indicates that disruption is blocked due to budget constraints or other conditions | "No allowed disruptions due to blocking budget" | 
| DisruptionLaunching | Normal | NodeClaim | Emitted when launching a replacement NodeClaim during disruption | "Launching NodeClaim: Drift" | 
| DisruptionTerminating | Normal | Node, NodeClaim | Emitted when terminating nodes/nodeclaims during disruption | "Disrupting Node: Drift" | 
| DisruptionWaitingReadiness | Normal | NodeClaim | Emitted when waiting for readiness before continuing disruption | "Waiting on readiness to continue disruption" | 
| Unconsolidatable | Normal | Node, NodeClaim | Indicates a node cannot be consolidated due to constraints | Various messages about why consolidation is blocked | 
| FailedScheduling | Warning | Pod | Emitted when a pod fails to schedule | "Failed to schedule pod, {error details}" | 
| NoCompatibleInstanceTypes | Warning | NodePool | Emitted when no instance types match NodePool requirements | "NodePool requirements filtered out all compatible available instance types" | 
| Nominated | Normal | Pod | Emitted when a pod is nominated to a node/nodeclaim | "Pod should schedule on: nodeclaim/{name}" | 
| NodeRepairBlocked | Warning | Node, NodeClaim, NodePool | Emitted when node repair is blocked | Various reasons why repair is blocked |
| Disrupted | Normal | Pod | Emitted when a pod is deleted to accommodate node termination | "Deleting the pod to accommodate the terminationTime..." | 
| Evicted | Normal | Pod | Emitted when a pod is evicted during node termination | "Evicted pod: {reason}" | 
| FailedDraining | Warning | Node | Emitted when node draining fails | "Failed to drain node, {error}" | 
| TerminationGracePeriodExpiring | Warning | Node, NodeClaim | Emitted when termination grace period is about to expire | "All pods will be deleted by {time}" | 
| AwaitingVolumeDetachment | Normal | Node | Emitted when waiting for volume attachments to be deleted | "Awaiting deletion of bound volumeattachments ({names})" | 
| FailedConsistencyCheck | Warning | NodeClaim | Emitted when NodeClaim consistency checks fail | Various consistency check failure messages | 
| InsufficientCapacityError | Warning | NodeClaim | Emitted when there's insufficient capacity to launch a NodeClaim | "NodeClaim {name} event: {error details}" | 
| UnregisteredTaintMissing | Warning | NodeClaim | Emitted when the unregistered taint is missing from a NodeClaim | "Missing karpenter.sh/unregistered taint which prevents registration related race conditions" | 
| NodeClassNotReady | Warning | NodeClaim | Emitted when the NodeClass is not ready | "NodeClaim {name} event: {error details}" |
| `SpotInterrupted` | Warning | Node, NodeClaim | When AWS sends a spot interruption warning for an instance | "Spot interruption warning was triggered" |
| `SpotRebalanceRecommendation` | Normal | Node, NodeClaim | When AWS recommends rebalancing spot instances | "Spot rebalance recommendation was triggered" |
| `InstanceStopping` | Warning | Node, NodeClaim | When an EC2 instance receives a stopping signal | "Instance is stopping" |
| `InstanceTerminating` | Warning | Node, NodeClaim | When an EC2 instance is being terminated | "Instance is terminating" |
| `InstanceUnhealthy` | Warning | Node, NodeClaim | When AWS reports an instance as unhealthy | "An unhealthy warning was triggered for the instance" |
| `TerminatingOnInterruption` | Warning | Node, NodeClaim | When Karpenter initiates termination due to interruption | "Interruption triggered termination for the NodeClaim/Node" |
| `WaitingOnNodeClaimTermination` | Normal | EC2NodeClass | When EC2NodeClass deletion is waiting for NodeClaims to terminate | "Waiting on NodeClaim termination for [names]" |
| *(No Reason)* | Warning | NodePool | When Karpenter cannot resolve the NodeClass for a NodePool | "Failed resolving NodeClass" |
| *(No Reason)* | Warning | NodeClaim | When Karpenter cannot resolve the NodeClass for a NodeClaim | "Failed resolving NodeClass" |
| *(No Reason)* | Warning | NodeClaim | When spot instance launch fails due to missing service-linked role permissions | "Attempted to launch a spot instance but failed due to \"AuthFailure.ServiceLinkedRoleCreationNotPermitted\"" |


## Troubleshooting Commands

View these events using kubectl:

```bash
# View all events sorted by timestamp
kubectl get events --sort-by='.lastTimestamp'

# View events for NodeClaims
kubectl get events --field-selector involvedObject.kind=NodeClaim

# View events for Nodes
kubectl get events --field-selector involvedObject.kind=Node

# View events for EC2NodeClass
kubectl get events --field-selector involvedObject.kind=EC2NodeClass

# View events for NodePools
kubectl get events --field-selector involvedObject.kind=NodePool

# Filter by event type
kubectl get events --field-selector type=Warning
kubectl get events --field-selector type=Normal

# View events for a specific resource
kubectl get events --field-selector involvedObject.name=<resource-name>
