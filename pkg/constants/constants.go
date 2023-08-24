package constants

const (
	IngressClassName = "frp"
)

const (
	AnnotationIngressClass = "kubernetes.io/ingress.class"

	AnnotationHostHeaderRewrite = "frpro.kubernetes.io/host-header-rewrite"
	AnnotationHeaderXFromWhere  = "frpro.kubernetes.io/header-x-from-where"
	AnnotationBackendProtocol   = "frpro.kubernetes.io/backend-protocol"
)
