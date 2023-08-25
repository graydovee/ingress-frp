package constants

const (
	IngressClassName = "frp"
)

const (
	AnnotationIngressClass = "kubernetes.io/ingress.class"

	AnnotationHostHeaderRewrite = "frp.kubernetes.io/host-header-rewrite"
	AnnotationHeaderXFromWhere  = "frp.kubernetes.io/header-x-from-where"
	AnnotationBackendProtocol   = "frp.kubernetes.io/backend-protocol"
)
