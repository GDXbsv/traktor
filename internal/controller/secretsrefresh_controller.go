// Package controller contains the logic for reconciling SecretsRefresh objects.
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	traktorv1alpha1 "github.com/GDXbsv/traktor/api/v1alpha1"
)

// SecretsRefreshReconciler reconciles a SecretsRefresh object
type SecretsRefreshReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=traktor.gdxcloud.net,resources=secretsrefreshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=traktor.gdxcloud.net,resources=secretsrefreshes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=traktor.gdxcloud.net,resources=secretsrefreshes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// The req.Name contains the Secret's name and req.Namespace contains the Secret's namespace
func (r *SecretsRefreshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// This reconcile is triggered by a Secret change (via SetupWithManager Watches)
	// The Secret has already been filtered by namespace and label selectors
	// req.Name contains the Secret's name, req.Namespace contains the Secret's namespace

	secretNamespace := req.Namespace
	secretName := req.Name

	// Safety check: Skip if this is the operator's own namespace to prevent self-restart loop
	operatorNamespace := os.Getenv("POD_NAMESPACE")
	if operatorNamespace == "" {
		operatorNamespace = "traktor-system" // fallback to default
	}

	if secretNamespace == operatorNamespace {
		logger.Info("Skipping operator's own namespace to prevent self-restart", "namespace", secretNamespace)
		return ctrl.Result{}, nil
	}

	logger.Info("Secret changed, filtering deployments that use this secret",
		"secret", secretName,
		"namespace", secretNamespace)

	// List all deployments in the namespace
	deploymentList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deploymentList, client.InNamespace(secretNamespace)); err != nil {
		logger.Error(err, "Failed to list deployments", "namespace", secretNamespace)
		return ctrl.Result{}, err
	}

	// Filter and restart only deployments that use the changed secret
	restartedCount := 0
	for i := range deploymentList.Items {
		deployment := &deploymentList.Items[i]

		// Check if deployment uses the changed secret
		if !r.deploymentUsesSecret(deployment, secretName) {
			continue
		}

		if err := r.restartDeployment(ctx, deployment); err != nil {
			logger.Error(err, "Failed to restart deployment",
				"deployment", deployment.Name,
				"namespace", deployment.Namespace)
			continue
		}

		logger.Info("Deployment restarted",
			"deployment", deployment.Name,
			"namespace", deployment.Namespace)
		restartedCount++
	}

	logger.Info("Completed deployment restart",
		"secret", secretName,
		"namespace", secretNamespace,
		"restartedCount", restartedCount,
		"totalDeployments", len(deploymentList.Items))

	return ctrl.Result{}, nil
}

// restartDeployment restarts a deployment using Strategic Merge Patch,
// similar to 'kubectl rollout restart deployment'
func (r *SecretsRefreshReconciler) restartDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	// Create a patch that adds/updates the restartedAt annotation
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"traktor.gdxcloud.net/restartedAt": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	// Apply the patch using StrategicMergePatchType
	return r.Patch(ctx, deployment, client.RawPatch(types.StrategicMergePatchType, patchBytes))
}

// deploymentUsesSecret checks if a deployment references the specified secret
// in volumes, environment variables, or envFrom
func (r *SecretsRefreshReconciler) deploymentUsesSecret(deployment *appsv1.Deployment, secretName string) bool {
	podSpec := &deployment.Spec.Template.Spec

	// Check volumes
	for _, volume := range podSpec.Volumes {
		if volume.Secret != nil && volume.Secret.SecretName == secretName {
			return true
		}
	}

	// Check all containers (init and regular)
	allContainers := append([]corev1.Container{}, podSpec.InitContainers...)
	allContainers = append(allContainers, podSpec.Containers...)

	for _, container := range allContainers {
		// Check envFrom
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil && envFrom.SecretRef.Name == secretName {
				return true
			}
		}

		// Check env
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				if env.ValueFrom.SecretKeyRef.Name == secretName {
					return true
				}
			}
		}
	}

	// Check ephemeral containers separately (different type)
	for _, container := range podSpec.EphemeralContainers {
		// Check envFrom
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil && envFrom.SecretRef.Name == secretName {
				return true
			}
		}

		// Check env
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				if env.ValueFrom.SecretKeyRef.Name == secretName {
					return true
				}
			}
		}
	}

	// Check imagePullSecrets
	for _, imagePullSecret := range podSpec.ImagePullSecrets {
		if imagePullSecret.Name == secretName {
			return true
		}
	}

	return false
}

// getFilteredNamespaces returns namespaces that match the selector
func (r *SecretsRefreshReconciler) getFilteredNamespaces(ctx context.Context, sr *traktorv1alpha1.SecretsRefresh) ([]corev1.Namespace, error) {
	namespaceList := &corev1.NamespaceList{}

	// If no selector is specified, return all namespaces
	if sr.Spec.NamespaceSelector == nil {
		if err := r.List(ctx, namespaceList); err != nil {
			return nil, err
		}
		return namespaceList.Items, nil
	}

	// Convert label selector to labels.Selector
	selector, err := metav1.LabelSelectorAsSelector(sr.Spec.NamespaceSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid namespace selector: %w", err)
	}

	// List all namespaces
	if err := r.List(ctx, namespaceList); err != nil {
		return nil, err
	}

	// Filter namespaces by selector
	var filteredNamespaces []corev1.Namespace
	for _, ns := range namespaceList.Items {
		if selector.Matches(labels.Set(ns.Labels)) {
			filteredNamespaces = append(filteredNamespaces, ns)
		}
	}

	return filteredNamespaces, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretsRefreshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create predicates to filter only real secret updates
	secretPredicates := predicate.Funcs{
		// Ignore Create events (including initial cache sync on controller restart)
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		// Only process Update events where data actually changed
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSecret, oldOk := e.ObjectOld.(*corev1.Secret)
			newSecret, newOk := e.ObjectNew.(*corev1.Secret)

			if !oldOk || !newOk {
				return false
			}

			// Check if the secret data or stringData actually changed
			// This prevents reconciliation on metadata-only updates
			oldDataHash := hashSecretData(oldSecret)
			newDataHash := hashSecretData(newSecret)

			return oldDataHash != newDataHash
		},
		// Ignore Delete events - we don't need to restart deployments when secrets are deleted
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		// Process Generic events (can happen with informer resync)
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&traktorv1alpha1.SecretsRefresh{}).
		// Watch for changes to Secrets in all namespaces with predicates
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findSecretsRefreshForSecret),
			builder.WithPredicates(secretPredicates),
		).
		Named("secretsrefresh").
		Complete(r)
}

// hashSecretData creates a hash of secret data for comparison
func hashSecretData(secret *corev1.Secret) string {
	if secret == nil {
		return ""
	}

	// Combine all data keys and values in a deterministic way
	keys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.Write(secret.Data[k])
		b.WriteString(";")
	}

	return b.String()
}

// findSecretsRefreshForSecret maps a Secret to SecretsRefresh objects that should watch it
func (r *SecretsRefreshReconciler) findSecretsRefreshForSecret(ctx context.Context, secret client.Object) []ctrl.Request {
	logger := log.FromContext(ctx)

	// List all SecretsRefresh objects
	srList := &traktorv1alpha1.SecretsRefreshList{}
	if err := r.List(ctx, srList); err != nil {
		logger.Error(err, "Failed to list SecretsRefresh objects")
		return []ctrl.Request{}
	}

	requests := make([]ctrl.Request, 0, len(srList.Items))
	for _, sr := range srList.Items {
		// Check if this secret's namespace matches the SecretsRefresh namespace selector
		namespaces, err := r.getFilteredNamespaces(ctx, &sr)
		if err != nil {
			logger.Error(err, "Failed to get filtered namespaces", "secretsRefresh", sr.Name)
			continue
		}

		// Check if secret's namespace is in filtered list
		namespaceMatches := false
		for _, ns := range namespaces {
			if ns.Name == secret.GetNamespace() {
				namespaceMatches = true
				break
			}
		}

		if !namespaceMatches {
			continue
		}

		// Check if secret matches the secret selector
		if sr.Spec.SecretSelector != nil {
			selector, err := metav1.LabelSelectorAsSelector(sr.Spec.SecretSelector)
			if err != nil {
				logger.Error(err, "Invalid secret selector", "secretsRefresh", sr.Name)
				continue
			}
			if !selector.Matches(labels.Set(secret.GetLabels())) {
				continue
			}
		}

		// This SecretsRefresh should be reconciled
		// Pass the Secret's name and namespace properly
		requests = append(requests, ctrl.Request{
			NamespacedName: client.ObjectKey{
				Name:      secret.GetName(),      // Secret's name
				Namespace: secret.GetNamespace(), // Secret's namespace
			},
		})
	}

	return requests
}
