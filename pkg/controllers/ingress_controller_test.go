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
	"time"
)

var YamlIngressStr = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gitea-ingress
  namespace: default
  annotations:
    frp.kubernetes.io/basic-auth: "username:password"
spec:
  ingressClassName: frp
  tls:
    - hosts:
        - gitea.example.com
      secretName: gitea-tls
  rules:
    - host: gitea.example.com
      http:
        paths:
          - path: /
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

var YamlSecretStr = `
apiVersion: v1
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURwakNDQW82Z0F3SUJBZ0lKQUpWd1pBK3VtenMwTUEwR0NTcUdTSWIzRFFFQkN3VUFNSUdPTVFzd0NRWUQKVlFRR0V3SmpiakVSTUE4R0ExVUVDQXdJVTJoaGJtZG9ZV2t4RVRBUEJnTlZCQWNNQ0ZOb1lXNW5hR0ZwTVJFdwpEd1lEVlFRS0RBaG5jbUY1Wkc5MlpURVFNQTRHQTFVRUN3d0hhVzVuY21WemN6RVFNQTRHQTFVRUF3d0hjV0ZsCmNpNWpiakVpTUNBR0NTcUdTSWIzRFFFSkFSWVRaM0poZVdSdmRtVmxRR2R0WVdsc0xtTnZiVEFlRncweU5EQXgKTURJeE1ETTBNVEJhRncwek16RXlNekF4TURNME1UQmFNRlF4Q3pBSkJnTlZCQVlUQWtOT01SRXdEd1lEVlFRSQpEQWhUYUdGdVoyaGhhVEVSTUE4R0ExVUVCd3dJVTJoaGJtZG9ZV2t4RFRBTEJnTlZCQW9NQkhGaFpYSXhFREFPCkJnTlZCQU1NQjNGaFpYSXVZMjR3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRRFcKd3lvWkFSdkdSSlVqRjN2eGlrVVZNakppSUF6S0NrdSt5T0dzZml3ZHk2a3YrMjFoTnhnU05zYlM1RzRpc3o1UApCQlRoNnZPWlJWQnZvRWdXWmQ2cWFBVzY3S2JTR0llMGdNL2FiZkZrbkJWS1VwWlV5d2dONHFUU2I5bm5nMC9LCkNudVU3d3o2QjJ2dUFoU012ZkRQOElBVnpCcWdGRDc0RzdRdEpReVh3dlBYcGkvcGYzTVNEYk14enNyY0lwRmwKaHRZUFJuSjhxeGxZdVJuUzB1SnEyVFRCL01lMjY5ems1VEdaK1M3U1B4WG4vRHNVUWNoYzhaK1dacVByalRsVAp5SGhFWHlQTytnTEZSZkZSV2YyNDVxUVBHWjNFRlhIek81cHdJVGsyMkI5aEZ0c21FbmhIb3lWS3RwK3R0RmJIClNLdmFPZjc0WlZkWktBS3YxRHpkQWdNQkFBR2pRREErTUFzR0ExVWREd1FFQXdJRWNEQVRCZ05WSFNVRUREQUsKQmdnckJnRUZCUWNEQVRBYUJnTlZIUkVFRXpBUmdnbHNiMk5oYkdodmMzU0hCSDhBQUFFd0RRWUpLb1pJaHZjTgpBUUVMQlFBRGdnRUJBR04zNWt1RmhzSHQ2UUVBZjN0VmwwOFhiVGRmeko1RjBvS2l3TjAvcUZkZloyYXdEMUVlCjhSNHlpbkZaOXNKNjF2TWRTeWlieEliK3REYVJEbHJhUlVpc2dzNDhxaWJ1eldqZFZKaWpSNllIQU42bTVXZ3UKczF5UFd2TjFmUG9LKzVmbjg0MXlSc2Rra2FmYW9jTzZDelRxZ3pxS2dTWWV1bTNZVWZaeGR6T28zbVd4REh1OQorbUVUb1Q2dEc5N2hJdnRiTWRFazlNWGRXNlUySW8zcVJlcDdBV3BOV1lKa0tqNmlFbU91Q25zeHE2c1gyQXFmCktaL2YrVnk3bXYyRkUrQWdCeE5PSDNwTFlTSEk0Tk81ZmNIVXVSaGhmS1RxSWdMRUJ6N0tKZUVDS3lZaUJXREUKb1NLZy8xTU94cU1nQm9lNnFIbWd3enc1bnFTWWprR00zSDQ9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
  tls.key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBMXNNcUdRRWJ4a1NWSXhkNzhZcEZGVEl5WWlBTXlncEx2c2pockg0c0hjdXBML3R0CllUY1lFamJHMHVSdUlyTStUd1FVNGVyem1VVlFiNkJJRm1YZXFtZ0Z1dXltMGhpSHRJRFAybTN4Wkp3VlNsS1cKVk1zSURlS2swbS9aNTROUHlncDdsTzhNK2dkcjdnSVVqTDN3ei9DQUZjd2FvQlErK0J1MExTVU1sOEx6MTZZdgo2WDl6RWcyek1jN0szQ0tSWlliV0QwWnlmS3NaV0xrWjB0TGlhdGswd2Z6SHR1dmM1T1V4bWZrdTBqOFY1L3c3CkZFSElYUEdmbG1hajY0MDVVOGg0UkY4anp2b0N4VVh4VVZuOXVPYWtEeG1keEJWeDh6dWFjQ0U1TnRnZllSYmIKSmhKNFI2TWxTcmFmcmJSV3gwaXIyam4rK0dWWFdTZ0NyOVE4M1FJREFRQUJBb0lCQUVCVHFOSmdnSjg4ZjZkSgpLM2pIdjdWL21aUEdvYzRLazNHTDNmeTZ0aUFlbG9pbXVMWjd1QndNaURVMjhyNDJEaDNBelRoMkZZejlOQUNiCmM3d3h1eVl6ampQVkdvcW5pazVJbnZtQUlPUFAxSmkwY0E3cDJYbS9QenRCQVhYVTRSdFZWSHJodDNOVXNjRlMKb2pFZDIzbU5RZkJGZUZ3bWRFNEFqbEZQWFp3KzNmb1JUWktwb1RaK2tDZ2dzOHhBOEN1cTJDRkUrdUczblo5dworZi9DeU52YUdvb1hmR2FUc3QxV0laQ0h0U1M3R0NCM3d0VGlVMGZCc01aQjQ2dml2M01OMUxVaklSYWw4MXRJCkdQaER1NHJFblNxRVRUWWprK3hJQXFxTVFzSzVSQ09FemQ1ZXhRTnZwWkJiUEQzUEdDUTFES3lOMmowd3ltbDgKSHI1K01JRUNnWUVBOTZSMkppTUhIMVYyNitIYXc5bnFEdjJFdTdyNHZhVzg5U0FTRHNXRGRYMk14QjgrMDJRQgpyaU5RZWR2eXU0QXRPTngrby9hMkJHVlRqOXpRbmEraHZsdkNXTno1bWJYNUt2WlNYbDRGZUd3eXovQVBGWHpuCk1YMUU3YlhTZ3RLaDdpL3ZObWdpbFMyL1dqaW9vaDVrRWMvQlh3U3BMaFJHdk5VODBqcitZMDBDZ1lFQTNnS2gKb2R6cFVYZER1U3FrM2kxVFY4RGlYUEZPVGVpbzdWUnkyLzFQdXZ2UVl1bnJ2N2J3b2Q3MmcvVzRpOVVhOUgrdwprSTl1Nm9RakxnSG5leU9OcE9MVUJYQnNYc0JjZnpkUWxMN1h5K3l3emN0bXc0RjIvc0xBVzhsUnBJbkZLbEp6Ckw3OGFPaUxHZlZHSGN2R1RPVkV2U1VRZlpoaDYxaW1XaGJ6K1Y5RUNnWUJjWDQ1cXYxb2l5QUJxRUg5SDJ2dEIKeUROQXk0ZUpSazlycUNEVVBiekJrS2wzWnFoS3RkMGlsYTJwSnZBdUhLdkJzQTNWSDJ2Wnkrb1ZtYXAvaDBudgo5YzVTMDJxUGVaK045UC9ZajMyKzQ2MDRmelZCTUt3VWU4UEFYN2c4Y0ZGU3hiS1hPdFRiaklyNkhuUll0TGxqCkkzbmY5WjhkdnhaN3paYTRYS1VUYVFLQmdRQ3VNNkpnUDlkVDlTRk95Z2RUem56Mi9vS2dLemdtS2NsamNFQXcKSGpQUnBJVi9GODNFUU9mUUhBT1N4OXhtM0hDcUtRZUNad25CT3EzZ0M5NTI0UTdqc3BockxDdmNyVlBtL3FCYwpGdU45UDl2N252NmpxWktWbEhzYmlueGxmelVXWUZ2QnUxSDVEQkJ6aE9XamE2cjU3cG9NQTBnZjlGVnVkbk9GCnZTWldBUUtCZ0RzR0ZGZzZHU0ZqOVVWdWFQdm94YVpHb29IZUF4dHhxaDl3SXdGK2FnUDYzRDhjQnZNaE45VUgKL201bGM4bWkrTFp2WTBiWHAraFc5OXMwWHdWdGdrNHp6TldIY1ExNU40TDk0V2xPV25iYW5BM3IxaG84ekxTRwowVW92ajdYS3FpNERPVkNsdXV1VzhzN0NQVE5COWJTTTJoVTVrRUI2S1ZSaTRYcE5leURjCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
kind: Secret
metadata:
  creationTimestamp: null
  name: gitea-tls
  namespace: default
type: kubernetes.io/tls
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
	var secret corev1.Secret
	if err := yaml.Unmarshal([]byte(YamlSecretStr), &secret); err != nil {
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
		WithObjects(&ingress, &service, &secret).
		Build()

	//frpCli := frp.NewClient("127.0.0.1", 7400, "admin", "admin")
	frpCli := frp.NewFakeSyncer()
	go func() {
		if err := frpCli.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
	}()
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
	time.Sleep(100 * time.Second)
}
