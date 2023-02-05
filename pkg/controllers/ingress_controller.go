package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/grydovee/ingress-frp/pkg/frp"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strconv"
)

type FrpIngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	clock.Clock

	FrpSyncer frp.Syncer
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=,resources=services,verbs=get

func (r *FrpIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling", "req", req)

	var ingress networkingv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &ingress); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cfgs := make(map[string]frp.Config)

	tlsMap, err := r.loadTlsSecrets(ctx, &ingress)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.PathType != nil && *path.PathType != networkingv1.PathTypePrefix {
				l.Info("only support pathType: Prefix")
				continue
			}
			if path.Backend.Service == nil {
				l.Info("service not defined")
				continue
			}
			key := types.NamespacedName{Name: path.Backend.Service.Name, Namespace: ingress.Namespace}
			var svc corev1.Service
			if err := r.Get(ctx, key, &svc); err != nil {
				if apierrors.IsNotFound(err) {
					l.Info("service not found", "key", key)
					continue
				}
				return ctrl.Result{Requeue: true}, err
			}
			switch svc.Spec.Type {
			case corev1.ServiceTypeClusterIP:
				cfg := frp.HttpConfig{}
				cfg.Host = rule.Host
				cfg.LocalIp = svcToDomain(&svc)
				cfg.LocalPort = strconv.Itoa(int(path.Backend.Service.Port.Number))
				cfg.Locations = path.Path
				name := fmt.Sprintf("%s/%s/%s", ingress.Namespace, ingress.Name, svc.Name)

				bytes := sha256.Sum256([]byte(name))
				cfg.Group = fmt.Sprintf("%x", bytes[:8])
				cfg.GroupKey = fmt.Sprintf("%x", bytes[:])

				if tls, ok := tlsMap[rule.Host]; ok && path.Path == "/" {
					cfgs[name+":https"] = &frp.HttpsReverseProxyConfig{
						HttpConfig: cfg,
						TlsCrt:     tls.crtBase64,
						TlsKey:     tls.keyBase64,
					}
				}
				cfgs[name+":http"] = &cfg
			default:
				l.Info("unsupported service type", "key", key)
			}
		}
	}

	l.Info("update frp config", "cfgs", fmt.Sprintf("%v", cfgs))
	r.FrpSyncer.SetProxies(req.String(), cfgs)
	return ctrl.Result{}, nil
}

func (r *FrpIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

	if r.FrpSyncer == nil {
		r.FrpSyncer = frp.NewFakeSyncer()
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				ingress, ok := event.Object.(*networkingv1.Ingress)
				if !ok {
					return false
				}
				return IngressMatch(ingress)
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				ingress, ok := deleteEvent.Object.(*networkingv1.Ingress)
				if !ok {
					return false
				}
				return IngressMatch(ingress)
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				ingressNew, ok := updateEvent.ObjectNew.(*networkingv1.Ingress)
				if !ok {
					return false
				}
				ingressOld, ok := updateEvent.ObjectOld.(*networkingv1.Ingress)
				if !ok {
					return false
				}
				return IngressMatch(ingressNew) || IngressMatch(ingressOld)
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				ingress, ok := genericEvent.Object.(*networkingv1.Ingress)
				if !ok {
					return false
				}
				return IngressMatch(ingress)
			},
		})).
		Complete(r)
}

type tlsCert struct {
	crtBase64 string
	keyBase64 string
}

func (r *FrpIngressReconciler) loadTlsSecrets(ctx context.Context, ingress *networkingv1.Ingress) (map[string]tlsCert, error) {
	secrets := make(map[string]tlsCert)
	for _, tls := range ingress.Spec.TLS {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: tls.SecretName, Namespace: ingress.Namespace}, secret); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		var t tlsCert
		t.crtBase64 = base64.StdEncoding.EncodeToString(secret.Data["tls.crt"])
		t.keyBase64 = base64.StdEncoding.EncodeToString(secret.Data["tls.key"])
		for _, host := range tls.Hosts {
			secrets[host] = t
		}
	}
	return secrets, nil
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
