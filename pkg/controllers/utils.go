package controllers

import (
	"github.com/grydovee/ingress-frp/pkg/constants"
	networkingv1 "k8s.io/api/networking/v1"
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
