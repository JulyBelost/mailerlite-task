package controllers

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"

    emailv1 "github.com/example/email-operator/api/v1"
    _ "github.com/go-logr/logr"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    _ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"
)

// EmailReconciler reconciles a Email object
type EmailReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=email.example.com,resources=emails,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=email.example.com,resources=emails/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=email.example.com,resources=emails/finalizers,verbs=update

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
    err = r.Get(ctx, client.ObjectKey{Name: email.Spec.SenderConfigName, Namespace: email.Namespace}, emailSenderConfig)
    if err != nil {
        logger.Error(err, "Failed to get EmailSenderConfig", "EmailSenderConfigName", email.Spec.SenderConfigName)
        email.Status.Status = "Failed"
        email.Status.Error = "Failed to get EmailSenderConfig"
        r.Status().Update(ctx, email)
        return ctrl.Result{}, nil
    }

    // Send email using MailerSend API
    mailerSendAPI := "https://api.mailersend.com/v1/email"
    requestBody, _ := json.Marshal(map[string]interface{}{
        "from": map[string]string{
            "email": emailSenderConfig.Spec.From,
        },
        "to": []map[string]string{
            {
                "email": email.Spec.To,
            },
        },
        "subject": email.Spec.Subject,
        "html":    email.Spec.Body,
    })

    httpReq, err := http.NewRequest("POST", mailerSendAPI, bytes.NewBuffer(requestBody))
    if err != nil {
        logger.Error(err, "Failed to create request")
        email.Status.Status = "Failed"
        email.Status.Error = "Failed to create request"
        r.Status().Update(ctx, email)
        return ctrl.Result{}, nil
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+emailSenderConfig.Spec.ApiToken)

    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        logger.Error(err, "Failed to send email")
        email.Status.Status = "Failed"
        email.Status.Error = "Failed to send email"
        r.Status().Update(ctx, email)
        return ctrl.Result{}, nil
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK {
        logger.Error(fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body), "Failed to send email")
        email.Status.Status = "Failed"
        email.Status.Error = string(body)
        r.Status().Update(ctx, email)
        return ctrl.Result{}, nil
    }

    // Update the Email status
    var responseMap map[string]interface{}
    if err := json.Unmarshal(body, &responseMap); err != nil {
        logger.Error(err, "Failed to unmarshal response")
        email.Status.Status = "Failed"
        email.Status.Error = "Failed to unmarshal response"
        r.Status().Update(ctx, email)
        return ctrl.Result{}, nil
    }

    messageID := responseMap["message_id"].(string)
    email.Status.Status = "Sent"
    email.Status.MessageID = messageID
    r.Status().Update(ctx, email)

    logger.Info("Email sent successfully", "Email", email.Name, "MessageID", messageID)
    return ctrl.Result{}, nil
}

func (r *EmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&emailv1.Email{}).
        Complete(r)
}
