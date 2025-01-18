package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainName string `json:"domainName"`
		WebsiteID  string `json:"websiteId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	typeName := strings.Split(body.DomainName, ".")[0]

	f := false

	podConfig := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       map[string]string{"app": typeName, "website-id": body.WebsiteID},
			Namespace:    K8sClient.Namespace,
			GenerateName: "pod-",
			UID:          types.UID(uuid.New().String()),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "container-0",
					Image: "ghcr.io/tyrelchambers/reddex-custom-website:latest",
					Ports: []corev1.ContainerPort{
						{
							Name:          "port-0",
							ContainerPort: 8000,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								Optional: &f,
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "reddex-custom-secrets",
								},
							},
						},
					},
					ImagePullPolicy: corev1.PullAlways,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: &f,
						Privileged:               &f,
						ReadOnlyRootFilesystem:   &f,
						RunAsNonRoot:             &f,
					},
				},
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "ghrc",
				},
			},
		},
	}

	pathType := v1.PathTypePrefix

	deployParams := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-deploy", typeName),
			Labels: map[string]string{
				"app":        typeName,
				"website-id": body.WebsiteID,
				"workload.user.cattle.io/workloadselector": fmt.Sprintf("apps.deployment-dynamic-sites-%s", typeName),
			},
			Namespace: K8sClient.Namespace,
			UID:       types.UID(uuid.New().String()),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workload.user.cattle.io/workloadselector": fmt.Sprintf("apps.deployment-dynamic-sites-%s", typeName),
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%"},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        typeName,
						"website-id": body.WebsiteID,
						"workload.user.cattle.io/workloadselector": fmt.Sprintf("apps.deployment-dynamic-sites-%s", typeName),
					},
				},
				Spec: podConfig.Spec,
			},
		},
	}

	d, err := CreateDeployment(deployParams)

	if err != nil {
		log.Error("Failed to create deployment", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serviceParams := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "service-",
			Labels: map[string]string{
				"websiteId": body.WebsiteID,
				"app":       typeName,
			},
			Namespace: K8sClient.Namespace,
			Annotations: map[string]string{
				"workload.user.cattle.io/workloadselector": d.Annotations["workload.user.cattle.io/workloadselector"],
			},
			UID: types.UID(uuid.New().String()),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":        typeName,
				"website-id": body.WebsiteID,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       fmt.Sprintf("%s-port", typeName),
					Port:       8000,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 8000},
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	s, err := CreateService(serviceParams)

	if err != nil {
		log.Error("Failed to create service", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ingressParams := v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       map[string]string{"app": typeName, "website-id": body.WebsiteID},
			GenerateName: "ingress-",
			Annotations: map[string]string{
				"cert-manager.io/cluster-issuer": "letsencrypt-prod",
			},
			Namespace: K8sClient.Namespace,
			UID:       types.UID(uuid.New().String()),
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: body.DomainName,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: s.Name,
											Port: v1.ServiceBackendPort{
												Number: 8000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []v1.IngressTLS{
				{
					Hosts:      []string{body.DomainName},
					SecretName: fmt.Sprintf("letsencrypt-%s", body.DomainName),
				},
			},
		},
	}

	_, err = CreateIngress(ingressParams)

	if err != nil {
		log.Error("Failed to create ingress", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func Delete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainName string `json:"domainName"`
		WebsiteID  string `json:"websiteId"`
	}

	errored := false

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	typeName := strings.Split(body.DomainName, ".")[0]

	fDep, err := FindDeployment(typeName)

	if err != nil {
		log.Error("Failed to find deployment", "error", err)
		errored = true
	}

	if fDep != nil {
		err = DeleteDeployment(DeleteDeploymentParams{
			DeploymentName: fDep.Name,
		})

		if err != nil {
			log.Error("Failed to delete deployment", "error", err)
			errored = true
		}
		log.Info("Deleted deployment", "name", fDep.Name)

	}

	fIngress, err := FindIngress(typeName)
	if err != nil {
		log.Error("Failed to find ingress", "error", err)
		errored = true
	}

	if fIngress != nil {

		err = DeleteIngress(DeleteIngressParams{
			IngressName: fIngress.Name,
		})

		if err != nil {
			log.Error("Failed to delete ingress", "error", err)
			errored = true
		}

		log.Info("Deleted ingress", "name", fIngress.Name)

	}

	fService, err := FindService(typeName)
	if err != nil {
		log.Error("Failed to find service", "error", err)
		errored = true
	}

	if fService != nil {

		err = DeleteService(DeleteServiceParams{
			ServiceName: fService.Name,
		})

		if err != nil {
			log.Error("Failed to delete service", "error", err)
			errored = true
		}
	}

	if errored {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
