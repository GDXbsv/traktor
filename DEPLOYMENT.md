# Traktor Operator Deployment Guide

This guide explains how to deploy the Traktor operator to your Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (v1.31+)
- `kubectl` configured to access your cluster
- Docker or Podman for building images
- Access to a container registry (Docker Hub, GCR, ECR, etc.)

## Deployment Options

### Option 1: Quick Deploy (Recommended for Development/Testing)

This method deploys the operator directly using kustomize.

#### Step 1: Build and Push the Container Image

```bash
# Set your container registry and image tag
export IMG=<your-registry>/traktor:v0.0.1
# Example: export IMG=docker.io/myuser/traktor:v0.0.1

# Build the container image
make docker-build IMG=$IMG

# Push to your registry
make docker-push IMG=$IMG
```

#### Step 2: Install CRDs

```bash
# Install the CustomResourceDefinitions
make install
```

This creates the `SecretsRefresh` CRD in your cluster.

#### Step 3: Deploy the Operator

```bash
# Deploy the operator to the cluster
make deploy IMG=$IMG
```

This will:
- Create the `traktor-system` namespace
- Deploy the operator controller manager
- Set up RBAC (ServiceAccount, Role, RoleBinding, ClusterRole, ClusterRoleBinding)
- Configure the manager deployment

#### Step 4: Verify the Deployment

```bash
# Check the operator pod
kubectl get pods -n traktor-system

# Expected output:
# NAME                                          READY   STATUS    RESTARTS   AGE
# traktor-controller-manager-xxxxxxxxxx-xxxxx   1/1     Running   0          30s

# Check the CRD
kubectl get crd secretsrefreshes.apps.gdxcloud.net

# View operator logs
kubectl logs -n traktor-system -l control-plane=controller-manager -f
```

---

### Option 2: Generate Install Manifest (Recommended for Production)

This method generates a single YAML file containing all resources for easy version control and GitOps workflows.

#### Step 1: Build and Push the Image

```bash
export IMG=<your-registry>/traktor:v0.0.1
make docker-build IMG=$IMG
make docker-push IMG=$IMG
```

#### Step 2: Generate the Installation Manifest

```bash
# Generate dist/install.yaml
make build-installer IMG=$IMG
```

This creates `dist/install.yaml` containing:
- Namespace
- CRDs
- RBAC resources
- Deployment

#### Step 3: Review and Apply

```bash
# Review the generated manifest
cat dist/install.yaml

# Apply to your cluster
kubectl apply -f dist/install.yaml

# Or commit to your GitOps repo
git add dist/install.yaml
git commit -m "Add traktor operator v0.0.1"
```

---

### Option 3: Local Development (No Container Required)

Run the operator locally on your development machine while it manages your cluster.

```bash
# Install CRDs
make install

# Run the operator locally (connects to your kubeconfig cluster)
make run
```

**Note:** This is useful for:
- Testing changes quickly
- Debugging with a local debugger
- Development without building containers

Press `Ctrl+C` to stop the operator.

---

## Configuration

### Understanding the SecretsRefresh Custom Resource

The operator watches for changes in Secrets based on filters you define, then restarts Deployments in those namespaces.

#### Example 1: Watch All Secrets in Production Namespaces

```yaml
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: production-secrets
  namespace: default
spec:
  # Watch namespaces with 'environment: production' label
  namespaceSelector:
    matchLabels:
      environment: production
  
  # Optional: Only watch specific secrets
  secretSelector:
    matchLabels:
      auto-refresh: "true"
  
  refreshInterval: "5m"
```

Apply it:
```bash
kubectl apply -f secretsrefresh-production.yaml
```

#### Example 2: Simple Configuration - Watch One Namespace

```yaml
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: app-secrets
  namespace: default
spec:
  # Watch only the 'app-backend' namespace
  namespaceSelector:
    matchExpressions:
      - key: kubernetes.io/metadata.name
        operator: In
        values:
          - app-backend
  
  refreshInterval: "10m"
```

#### Example 3: Watch All Namespaces

```yaml
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: all-secrets
  namespace: default
spec:
  # No namespaceSelector = watch all namespaces
  # No secretSelector = watch all secrets
  refreshInterval: "15m"
```

### Labeling Your Resources

For the operator to watch your resources, label them appropriately:

#### Label Namespaces

```bash
# Label a namespace to be watched
kubectl label namespace my-app environment=production

# Label multiple namespaces
kubectl label namespace app-backend app-frontend team=platform
```

#### Label Secrets

```bash
# Label a secret to be watched
kubectl label secret my-app-secret -n my-app auto-refresh=true

# Create a secret with labels
kubectl create secret generic my-secret \
  --from-literal=key=value \
  -n my-app \
  --dry-run=client -o yaml | \
  kubectl label --local -f - auto-refresh=true --dry-run=client -o yaml | \
  kubectl apply -f -
```

---

## How It Works

1. **Watch Setup**: When you create a `SecretsRefresh` CR, the operator starts watching Secrets matching your filters
2. **Secret Change Detection**: When a watched Secret changes, the operator detects it
3. **Deployment Restart**: The operator restarts ALL Deployments in the namespace where the Secret changed by:
   - Adding/updating the annotation: `traktor.gdxcloud.net/restartedAt: <timestamp>`
   - This triggers a rolling restart of the pods

**Example Flow:**
```
1. Secret 'db-password' changes in namespace 'backend'
2. Operator detects the change (Secret matches filters)
3. Operator finds all Deployments in 'backend' namespace
4. Operator adds annotation to each Deployment's pod template
5. Kubernetes performs rolling restart of all pods
6. Pods now pick up the new secret values
```

---

## Verification

### Check Operator Status

```bash
# View operator logs
kubectl logs -n traktor-system deployment/traktor-controller-manager

# Check for errors
kubectl get events -n traktor-system --sort-by='.lastTimestamp'
```

### Test the Operator

1. **Create a test namespace and label it:**
```bash
kubectl create namespace test-traktor
kubectl label namespace test-traktor watch-secrets=true
```

2. **Create a SecretsRefresh CR:**
```bash
cat <<EOF | kubectl apply -f -
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: test-refresh
  namespace: default
spec:
  namespaceSelector:
    matchLabels:
      watch-secrets: "true"
  refreshInterval: "5m"
EOF
```

3. **Create a test deployment:**
```bash
kubectl create deployment nginx --image=nginx -n test-traktor
```

4. **Create a labeled secret:**
```bash
kubectl create secret generic test-secret \
  --from-literal=password=oldpass \
  -n test-traktor
```

5. **Update the secret and watch deployments restart:**
```bash
# Update the secret
kubectl create secret generic test-secret \
  --from-literal=password=newpass \
  -n test-traktor \
  --dry-run=client -o yaml | kubectl apply -f -

# Watch the deployment - you should see a rollout
kubectl rollout status deployment/nginx -n test-traktor

# Check the annotation
kubectl get deployment nginx -n test-traktor -o jsonpath='{.spec.template.metadata.annotations}'
# Should show: {"traktor.gdxcloud.net/restartedAt":"2024-01-15T10:30:00Z"}
```

---

## Uninstallation

### Remove the Operator

```bash
# Remove the operator deployment
make undeploy

# Remove the CRDs (this will DELETE all SecretsRefresh resources!)
make uninstall
```

Or if you used the install manifest:

```bash
kubectl delete -f dist/install.yaml
```

### Cleanup Test Resources

```bash
kubectl delete namespace test-traktor
```

---

## Troubleshooting

### Operator Pod Not Starting

```bash
# Check pod status
kubectl get pods -n traktor-system

# View pod events
kubectl describe pod -n traktor-system -l control-plane=controller-manager

# Check logs
kubectl logs -n traktor-system -l control-plane=controller-manager
```

### CRD Installation Failed

```bash
# Check if CRD exists
kubectl get crd secretsrefreshes.apps.gdxcloud.net

# Reinstall CRD
make install
```

### Deployments Not Restarting

Check the operator logs:
```bash
kubectl logs -n traktor-system deployment/traktor-controller-manager -f
```

Common issues:
- Secret doesn't match the `secretSelector` labels
- Namespace doesn't match the `namespaceSelector` labels
- RBAC permissions missing (check ClusterRole)

### View RBAC Permissions

```bash
# View what the operator can do
kubectl describe clusterrole traktor-manager-role

# Check if ServiceAccount has proper bindings
kubectl get clusterrolebinding | grep traktor
```

---

## Multi-Cluster Deployment

To deploy across multiple clusters:

```bash
# Build once
export IMG=<your-registry>/traktor:v0.0.1
make docker-build IMG=$IMG
make docker-push IMG=$IMG

# Deploy to each cluster
for cluster in prod-us prod-eu staging; do
  kubectl config use-context $cluster
  make deploy IMG=$IMG
done
```

---

## Security Considerations

1. **RBAC Permissions**: The operator requires cluster-wide read access to Secrets and Namespaces, and write access to Deployments
2. **Image Security**: Use a private registry for production deployments
3. **Network Policies**: Consider restricting operator network access
4. **Pod Security**: The operator runs as non-root user (65532)

---

## Upgrading

```bash
# Pull latest changes
git pull

# Build new version
export IMG=<your-registry>/traktor:v0.0.2
make docker-build IMG=$IMG
make docker-push IMG=$IMG

# Update CRDs (if changed)
make install

# Upgrade deployment
make deploy IMG=$IMG
```

---

## Support

- **GitHub Issues**: Report bugs or request features
- **Logs**: Always include operator logs when reporting issues
- **Version**: Check your version: `kubectl get deployment -n traktor-system traktor-controller-manager -o jsonpath='{.spec.template.spec.containers[0].image}'`
