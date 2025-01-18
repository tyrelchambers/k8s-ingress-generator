package k8s

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateIngress(opts v1.Ingress) (*v1.Ingress, error) {
	ctx := context.TODO()

	ingress, err := K8sClient.Client.NetworkingV1().Ingresses(K8sClient.Namespace).Create(ctx, &opts, metav1.CreateOptions{})

	if err != nil {
		log.Error("Failed to create ingress", "name", opts.Name, "error", err)
		return nil, err
	}

	log.Info("Created ingress", "name", ingress.Name)
	return ingress, nil
}

func CreateService(opts corev1.Service) (*corev1.Service, error) {
	ctx := context.TODO()
	serv, err := K8sClient.Client.CoreV1().Services(K8sClient.Namespace).Create(ctx, &opts, metav1.CreateOptions{})

	if err != nil {
		log.Error("Failed to create service", "name", opts.Name, "error", err)
		return nil, err
	}

	log.Info("Created service", "name", serv.Name)
	return serv, nil
}

func CreatePod(opts corev1.Pod) (*corev1.Pod, error) {
	ctx := context.TODO()
	pod, err := K8sClient.Client.CoreV1().Pods(K8sClient.Namespace).Create(ctx, &opts, metav1.CreateOptions{})

	if err != nil {
		log.Error("Failed to create pod", "name", opts.Name, "error", err)
		return nil, err
	}

	log.Info("Created pod", "name", pod.Name)
	return pod, nil
}

func CreateDeployment(opts appsv1.Deployment) (*appsv1.Deployment, error) {
	ctx := context.TODO()

	dep, err := K8sClient.Client.AppsV1().Deployments(K8sClient.Namespace).Create(ctx, &opts, metav1.CreateOptions{})

	if err != nil {
		log.Error("Failed to create deployment", "name", opts.Name, "error", err)
		return nil, err
	}

	log.Info("Created deployment", "name", dep.Name)
	return dep, nil
}

func FindDeployment(name string) (*appsv1.Deployment, error) {
	ctx := context.TODO()

	log.Info("Finding deployment", "name", name)

	dep, err := K8sClient.Client.AppsV1().Deployments(K8sClient.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	})

	fmt.Println(fmt.Sprintf("app=%s", name))

	if err != nil {
		return nil, err
	}

	if len(dep.Items) == 0 {
		return nil, errors.New("Deployment not found")
	}

	return &dep.Items[0], nil
}

type DeleteDeploymentParams struct {
	DeploymentName string
}

func DeleteDeployment(params DeleteDeploymentParams) error {
	err := K8sClient.Client.AppsV1().Deployments(K8sClient.Namespace).Delete(context.TODO(), params.DeploymentName, metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}

type DeleteIngressParams struct {
	IngressName string
}

func FindIngress(name string) (*v1.Ingress, error) {
	ctx := context.TODO()

	log.Info("Finding ingress", "name", name)

	ingress, err := K8sClient.Client.NetworkingV1().Ingresses(K8sClient.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	})

	if err != nil {
		return nil, err
	}

	if len(ingress.Items) == 0 {
		return nil, errors.New("ingress not found")
	}

	return &ingress.Items[0], nil
}

func DeleteIngress(params DeleteIngressParams) error {
	err := K8sClient.Client.NetworkingV1().Ingresses(K8sClient.Namespace).Delete(context.TODO(), params.IngressName, metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}

func FindService(name string) (*corev1.Service, error) {
	ctx := context.TODO()

	log.Info("Finding service", "name", name)

	serv, err := K8sClient.Client.CoreV1().Services(K8sClient.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	})
	if err != nil {
		return nil, err
	}

	if len(serv.Items) == 0 {
		return nil, errors.New("Service not found")
	}

	return &serv.Items[0], nil
}

type DeleteServiceParams struct {
	ServiceName string
}

func DeleteService(params DeleteServiceParams) error {
	err := K8sClient.Client.CoreV1().Services(K8sClient.Namespace).Delete(context.TODO(), params.ServiceName, metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}
