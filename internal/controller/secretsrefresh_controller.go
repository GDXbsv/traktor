// Package controller contains the logic for reconciling SecretsRefresh objects.
package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	traktorv1alpha1 "github.com/GDXbsv/traktor/api/v1alpha1"
)

// SecretsRefreshReconciler reconciles a SecretsRefresh object
type SecretsRefreshReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=apps.gdxcloud.net,resources=secretsrefreshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gdxcloud.net,resources=secretsrefreshes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.gdxcloud.net,resources=secretsrefreshes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecretsRefreshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// This reconcile is triggered by a Secret change (via SetupWithManager Watches)
	// The Secret has already been filtered by namespace and label selectors
	// We just need to restart all Deployments in the same namespace

	logger.Info("Secret changed, restarting deployments", "namespace", req.Namespace)

	// List all deployments in the namespace
	deploymentList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
		logger.Error(err, "Failed to list deployments", "namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	// Restart each deployment by updating the annotation
	restartedCount := 0
	for i := range deploymentList.Items {
		deployment := &deploymentList.Items[i]

		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}

		deployment.Spec.Template.Annotations["traktor.gdxcloud.net/restartedAt"] = time.Now().Format(time.RFC3339)

		if err := r.Update(ctx, deployment); err != nil {
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
		"namespace", req.Namespace,
		"restartedCount", restartedCount,
		"totalDeployments", len(deploymentList.Items))

	return ctrl.Result{}, nil
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&traktorv1alpha1.SecretsRefresh{}).
		// Watch for changes to Secrets in all namespaces
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findSecretsRefreshForSecret),
		).
		Named("secretsrefresh").
		Complete(r)
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

	var requests []ctrl.Request
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
		requests = append(requests, ctrl.Request{
			NamespacedName: client.ObjectKey{
				Name:      sr.Name,
				Namespace: sr.Namespace,
			},
		})
	}

	return requests
}
