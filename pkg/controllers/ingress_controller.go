package controllers

import (
	"context"
	"fmt"
	"github.com/grydovee/ingress-frp/pkg/frp"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	"net"
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

	FrpClient frp.Client
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
				localIp, err := toLocalIp(&svc)
				if err != nil {
					l.Info("can not get service ip", "key", key, "err", err)
					continue
				}
				cfg := &frp.HttpConfig{}
				cfg.Host = rule.Host
				cfg.LocalIp = localIp
				cfg.LocalPort = strconv.Itoa(int(path.Backend.Service.Port.Number))
				cfg.Locations = path.Path
				name := fmt.Sprintf("%s/%s/%s", ingress.Namespace, ingress.Name, svc.Name)

				cfgs[name] = cfg
			default:
				l.Info("unsupported service type", "key", key)
			}
		}
	}

	l.Info("update frp config", "cfgs", cfgs)
	configs, err := r.FrpClient.GetConfigs()
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if configs.Proxy == nil {
		configs.Proxy = make(map[string]frp.Config)
	}

	for name, config := range cfgs {
		configs.Proxy[name] = config
	}

	if err = r.FrpClient.SetConfig(configs); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.FrpClient.Reload(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func toLocalIp(service *corev1.Service) (string, error) {
	if service == nil {
		return "", fmt.Errorf("service is nil")
	}
	clusterIp := service.Spec.ClusterIP
	ip := net.ParseIP(clusterIp)
	if ip != nil {
		return ip.String(), nil
	}

	domain := fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
	lookupIPs, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}
	for _, p := range lookupIPs {
		if p.To4() != nil {
			return p.String(), nil
		}
	}
	for _, p := range lookupIPs {
		if p.To16() != nil {
			return p.String(), nil
		}
	}
	return "", fmt.Errorf("no address found for %s", domain)
}

func (r *FrpIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

	if r.FrpClient == nil {
		r.FrpClient = frp.NewFakeClient()
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
