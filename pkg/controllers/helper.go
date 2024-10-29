package controllers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/grydovee/ingress-frp/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"strconv"
)

func IngressMatch(ingress *networkingv1.Ingress) bool {
	if ingress == nil {
		return false
	}
	var ingressClassName string
	ingressClassName = ingress.Annotations[constants.AnnotationIngressClass]

	if ingress.Spec.IngressClassName != nil {
		if len(ingressClassName) == 0 {
			ingressClassName = *ingress.Spec.IngressClassName
		} else if *ingress.Spec.IngressClassName != ingressClassName {
			return false
		}
	}
	return ingressClassName == constants.IngressClassName
}

func GenerateGroup(name, proxyType string) (string, string) {
	hashKey := fmt.Sprintf("%s/%s", name, proxyType)
	bytes := sha256.Sum256([]byte(hashKey))
	return fmt.Sprintf("%x", bytes[:8]), fmt.Sprintf("%x", bytes[:])
}

func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func svcToDomain(service *corev1.Service) string {
	if service == nil {
		return ""
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
}

func getIngressPort(path *networkingv1.HTTPIngressPath, service *corev1.Service) (string, error) {
	if path == nil || service == nil {
		return "", fmt.Errorf("path or service is nil")
	}

	if port := path.Backend.Service.Port.Number; port != 0 {
		return strconv.Itoa(int(port)), nil
	}

	if name := path.Backend.Service.Port.Name; name != "" {
		for _, port := range service.Spec.Ports {
			if port.Name == name {
				return strconv.Itoa(int(port.Port)), nil
			}
		}
	}
	return "", fmt.Errorf("port not found")
}
