package controllers

import (
	"context"
	"fmt"
	"github.com/grydovee/ingress-frp/pkg/frp"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/clock"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var YamlIngressStr = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gitea-ingress
  namespace: default
spec:
  ingressClassName: frp
  tls:
    - hosts:
        - gitea.example.com
      secretName: gitea-tls
  rules:
    - host: 121.36.18.41
      http:
        paths:
          - path: /xxx
            pathType: Prefix
            backend:
              service:
                name: gitea
                port:
                  number: 3000
`

var YamlServiceStr = `
apiVersion: v1
kind: Service
metadata:
  name: gitea
  namespace: default
spec:
  selector:
    app.kubernetes.io/name: gitea
  type: ClusterIP
  clusterIP: 127.0.0.1
  ports:
  - port: 3000
    targetPort: 3000
    protocol: TCP
    name: http
`

func TestFrpIngressReconciler_Reconcile(t *testing.T) {
	var ingress networkingv1.Ingress
	if err := yaml.Unmarshal([]byte(YamlIngressStr), &ingress); err != nil {
		t.Fatal(err)
	}
	var service corev1.Service
	if err := yaml.Unmarshal([]byte(YamlServiceStr), &service); err != nil {
		t.Fatal(err)
	}
	scheme := runtime.NewScheme()
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	cli := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(&ingress, &service).
		Build()

	//frpCli := frp.NewClient("127.0.0.1", 7400, "admin", "admin")
	frpCli := frp.NewFakeSyncer()
	reconciler := &FrpIngressReconciler{
		Client:    cli,
		Scheme:    scheme,
		Clock:     clock.RealClock{},
		FrpSyncer: frpCli,
	}
	tests := []struct {
		req     controllerruntime.Request
		wantErr bool
	}{
		{
			req:     controllerruntime.Request{NamespacedName: client.ObjectKeyFromObject(&ingress)},
			wantErr: false,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			_, err := reconciler.Reconcile(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
