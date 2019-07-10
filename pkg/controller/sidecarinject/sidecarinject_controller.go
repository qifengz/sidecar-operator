/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sidecarinject

import (
	"context"
	"fmt"
	"sigs.k8s.io/yaml"
	shipv1 "github.com/sidecar-operator/pkg/apis/ship/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	_types "k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SidecarInject Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSidecarInject{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sidecarinject-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to SidecarInject
	err = c.Watch(&source.Kind{Type: &shipv1.SidecarInject{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSidecarInject{}

// ReconcileSidecarInject reconciles a SidecarInject object
type ReconcileSidecarInject struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SidecarInject object and makes changes based on the state read
// and what is in the SidecarInject.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ship.github.com,resources=sidecarinjects,verbs=get;;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ship.github.com,resources=sidecarinjects/status,verbs=get;update;patch
func (r *ReconcileSidecarInject) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the SidecarInject instance
	instance := &shipv1.SidecarInject{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// validating configmap data
	configMap := &corev1.ConfigMap{}
	sidecarTemplateConfigMapName := instance.Spec.SidecarTemplateConfigmapName
	if sidecarTemplateConfigMapName == "" {
		sidecarTemplateConfigMapName = "sidecar-templ-configmap"
	}
	operatorNamespace := request.Namespace
	err = r.Get(context.TODO(), _types.NamespacedName{Namespace: operatorNamespace, Name: sidecarTemplateConfigMapName}, configMap)
	if err != nil {
		// if not found, retry
		if errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	logf.Log.Info("get configMap", "configmap name", sidecarTemplateConfigMapName, "configmap namespace", operatorNamespace, "configMap.Data", configMap.Data)
	// validating
	replicas := instance.Spec.SidecarNum
	if replicas < 0 {
		return reconcile.Result{}, errors.NewBadRequest("num cannot be less than 0")
	}

	var sidecar *corev1.Container
	if sidecarTemplate, exists := configMap.Data["sidecar-template"]; exists {
		err = yaml.Unmarshal([]byte(sidecarTemplate), &sidecar)
		if err != nil {
			return reconcile.Result{}, errors.NewBadRequest("sidecar-template is invalid")
		}
	}

	// list objective pods
	pods := &corev1.PodList{}
	err = r.List(context.TODO(),
		client.InNamespace(instance.Namespace).
			MatchingLabels(instance.Spec.Selector),
		pods)

	if err != nil {
		logf.Log.Error(err, "pod list not found")
		return reconcile.Result{}, err
	}

	// delete pod to trigger updating
	for _, pod := range pods.Items {
		curCount := 0
		for _, container := range pod.Spec.Containers {
			if container.Name == sidecar.Name && container.Image == sidecar.Image {
				curCount += 1
			}
		}
		// reach to expected state, skip
		if curCount == replicas {
			return reconcile.Result{}, nil
		}

		os.Setenv("SIDECAR_CONFIGMAP_NAME", sidecarTemplateConfigMapName)
		configMap.Data["num"] = fmt.Sprintf("%d", instance.Spec.SidecarNum)
		err = r.Update(context.TODO(), configMap)
		if err != nil {
			logf.Log.Error(err, "update configmap data num failed")
			return reconcile.Result{}, err
		}

		// assume the pod owned by a deployment or statefulset controller, delete pod for recreating it
		err = r.Delete(context.TODO(), &pod, client.GracePeriodSeconds(30))
		if err != nil {
			logf.Log.Error(err, "delete pod for recreating pod failed")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
