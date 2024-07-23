package controllers

import (
    "context"
    "fmt"

    "github.com/go-logr/logr"
    emailv1 "github.com/example/email-operator/api/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"
)

// EmailSenderConfigReconciler reconciles a EmailSenderConfig object
type EmailSenderConfigReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=email.example.com,resources=emailsenderconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=email.example.com,resources=emailsenderconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=email.example.com,resources=emailsenderconfigs/finalizers,verbs=update

func (r *EmailSenderConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Fetch the EmailSenderConfig instance
    emailSenderConfig := &emailv1.EmailSenderConfig{}
    err := r.Get(ctx, req.NamespacedName, emailSenderConfig)
    if err != nil {
        if client.IgnoreNotFound(err) != nil {
            logger.Error(err, "Failed to get EmailSenderConfig")
        }
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Confirm email sending settings
    if emailSenderConfig.ObjectMeta.DeletionTimestamp.IsZero() {
        if !controllerutil.ContainsFinalizer(emailSenderConfig, "email.example.com/finalizer") {
            controllerutil.AddFinalizer(emailSenderConfig, "email.example.com/finalizer")
            err := r.Update(ctx, emailSenderConfig)
            if err != nil {
                logger.Error(err, "Failed to add finalizer to EmailSenderConfig")
                return ctrl.Result{}, err
            }
            logger.Info("Added finalizer to EmailSenderConfig", "EmailSenderConfig", emailSenderConfig.Name)
        }
    } else {
        if controllerutil.ContainsFinalizer(emailSenderConfig, "email.example.com/finalizer") {
            // Handle any cleanup logic here
            controllerutil.RemoveFinalizer(emailSenderConfig, "email.example.com/finalizer")
            err := r.Update(ctx, emailSenderConfig)
            if err != nil {
                logger.Error(err, "Failed to remove finalizer from EmailSenderConfig")
                return ctrl.Result{}, err
            }
            logger.Info("Removed finalizer from EmailSenderConfig", "EmailSenderConfig", emailSenderConfig.Name)
        }
        return ctrl.Result{}, nil
    }

    logger.Info("Reconciled EmailSenderConfig", "EmailSenderConfig", emailSenderConfig.Name)
    return ctrl.Result{}, nil
}

func (r *EmailSenderConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&emailv1.EmailSenderConfig{}).
        Complete(r)
}
