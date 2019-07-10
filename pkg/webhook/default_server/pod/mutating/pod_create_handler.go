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

package mutating

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	_types "k8s.io/apimachinery/pkg/types"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
	"sigs.k8s.io/yaml"
	"strconv"
)

func init() {
	webhookName := "inject"
	if HandlerMap[webhookName] == nil {
		HandlerMap[webhookName] = []admission.Handler{}
	}
	HandlerMap[webhookName] = append(HandlerMap[webhookName], &PodCreateHandler{})
	fmt.Printf("HandlerMap: %v", HandlerMap)
}

// PodCreateHandler handles Pod
type PodCreateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	Client client.Client

	// Decoder decodes objects
	Decoder types.Decoder
}

func (h *PodCreateHandler) mutatingPodFn(ctx context.Context, pod *corev1.Pod, container corev1.Container, replicas int) error {
	containerName := container.Name
	for i := 0; i < replicas; i++ {
		container.Name = fmt.Sprintf("%s-%d", containerName, i)
		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}
	return nil
}

var _ admission.Handler = &PodCreateHandler{}

// Handle handles admission requests.
func (h *PodCreateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	obj := &corev1.Pod{}
	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := obj.DeepCopy()
	configMap := &corev1.ConfigMap{}
	sidecarTemplateConfigMapName := os.Getenv("SIDECAR_CONFIGMAP_NAME")
	// default configmap
	if sidecarTemplateConfigMapName == "" {
		sidecarTemplateConfigMapName = "sidecar-templ-configmap"
	}
	// sidecar configmap namespace is same with pod's namespace
	configmapNamespace := req.AdmissionRequest.Namespace
	// logf.Log.Info("get operatorNamespace", "operatorNamespace", operatorNamespace)
	err = h.Client.Get(ctx, _types.NamespacedName{Namespace: configmapNamespace, Name: sidecarTemplateConfigMapName}, configMap)
	if err != nil {
		logf.Log.Error(err, "configMap not found", configmapNamespace, sidecarTemplateConfigMapName)
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	logf.Log.Info("Handle configMap.Data", "configMap.Data", configMap.Data)
	replicas := 0
	container := &corev1.Container{}
	if num, numExists := configMap.Data["num"]; numExists {
		replicas, err = strconv.Atoi(num)
		if err != nil {
			return admission.ErrorResponse(http.StatusBadRequest, err)
		}

		if replicas < 0 {
			return admission.ErrorResponse(http.StatusBadRequest, errors.NewBadRequest("num cannot be less than 0"))
		}

		if replicas > 1000 {
			return admission.ErrorResponse(http.StatusBadRequest, errors.NewBadRequest("num cannot be greater than 1000"))
		}

		if sidecarTemplate, exists := configMap.Data["sidecar-template"]; exists {
			err = yaml.Unmarshal([]byte(sidecarTemplate), container)
			if err != nil {
				return admission.ErrorResponse(http.StatusBadRequest, errors.NewBadRequest("sidecar-template content is invalid"))
			}
		}
	}

	logf.Log.Info("get sidecar container", "sidecar container", container)
	err = h.mutatingPodFn(ctx, copy, *container, replicas)
	logf.Log.Info("get pod spec containers", "pod containers", copy.Spec.Containers)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return admission.PatchResponse(obj, copy)
}

var _ inject.Client = &PodCreateHandler{}
//
// // InjectClient injects the client into the PodCreateHandler
func (h *PodCreateHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}

var _ inject.Decoder = &PodCreateHandler{}

// InjectDecoder injects the decoder into the PodCreateHandler
func (h *PodCreateHandler) InjectDecoder(d types.Decoder) error {
	h.Decoder = d
	return nil
}
