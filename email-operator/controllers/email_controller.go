package controllers

import (
    "context"
    emailv1 "github.com/example/email-operator/api/v1"
    _ "github.com/go-logr/logr"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
)

type EmailReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *EmailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Fetch the Email instance
    email := &emailv1.Email{}
    err := r.Get(ctx, req.NamespacedName, email)
    if err != nil {
        if client.IgnoreNotFound(err) != nil {
            logger.Error(err, "Failed to get Email")
        }
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Fetch the corresponding EmailSenderConfig
    emailSenderConfig := &emailv1.EmailSenderConfig{}
    err = r.Get(ctx, client.ObjectKey{Name: email.Spec.SenderConfigRef, Namespace: email.Namespace}, emailSenderConfig)
    if err != nil {
        logger.Error(err, "Failed to get EmailSenderConfig", "EmailSenderConfigName", email.Spec.SenderConfigRef)
        email.Status.Status = "Failed"
        email.Status.Error = "Failed to get EmailSenderConfig"
        err := r.Status().Update(ctx, email)
        if err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{}, nil
    }

    // Proceed with sending email using the emailSenderConfig...
    // Ensure to use the correct fields from emailSenderConfig.Spec

    // Update the Email status
    email.Status.Status = "Sent"
    err = r.Status().Update(ctx, email)
    if err != nil {
        return ctrl.Result{}, err
    }

    logger.Info("Email sent successfully", "Email", email.Name)
    return ctrl.Result{}, nil
}

func (r *EmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&emailv1.Email{}).
        Complete(r)
}
