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

	frpClient frp.Client
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingress,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingress/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=,resources=service,verbs=get

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
				if len(svc.Spec.ClusterIP) == 0 {
					l.Info("clusterIp not found", "key", key)
					continue
				}
				cfg := &frp.HttpConfig{}
				cfg.Host = rule.Host
				cfg.LocalIp = svc.Spec.ClusterIP
				cfg.LocalPort = strconv.Itoa(int(path.Backend.Service.Port.Number))
				cfg.Locations = path.Path
				name := fmt.Sprintf("%s/%s/%s", ingress.Namespace, ingress.Name, svc.Name)

				cfgs[name] = cfg
			default:
				l.Info("unsupported service type", "key", key)
			}
		}
	}

	configs, err := r.frpClient.GetConfigs()
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if configs.Proxy == nil {
		configs.Proxy = make(map[string]frp.Config)
	}

	for name, config := range cfgs {
		configs.Proxy[name] = config
	}

	if err = r.frpClient.SetConfig(configs); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.frpClient.Reload(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
func (r *FrpIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

	if r.frpClient == nil {
		r.frpClient = frp.NewFakeClient()
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
