/*
Copyright 2024.

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

package controller

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	myv1alpha1 "podinfo/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=my.api.group,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=my.api.group,resources=myappresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=my.api.group,resources=myappresources/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("myAppResource", req.NamespacedName)
	myAppResource := &myv1alpha1.MyAppResource{}
	err := r.Get(ctx, req.NamespacedName, myAppResource)
	if err != nil {
		if errors.IsNotFound(err) {
			// object not found, could have been deleted after
			// reconcile request, hence don't requeue
			return ctrl.Result{}, nil
		}

		// error reading the object, requeue the request
		return ctrl.Result{}, err
	}
	log.Info("reconciling deployment")
	// Deployment Logic
	deployment := NewDeploymentForCr(myAppResource)
	err = ctrl.SetControllerReference(myAppResource, deployment, r.Scheme)
	if err != nil {
		// requeue with error
		return ctrl.Result{}, err
	}

	foundDeployment := &appsv1.Deployment{}
	// try to see if the deployment already exists
	err = r.Get(ctx, req.NamespacedName, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		// does not exist, create a deployment
		err = r.Create(ctx, deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Successfully created a deployment
		log.Info("deployment created successfully", "name", deployment.Name)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "cannot create deployment")
		return ctrl.Result{}, err
	} else if err == nil {
		// Check if ReplicaCount was changed
		if foundDeployment.Spec.Replicas != deployment.Spec.Replicas {
			foundDeployment.Spec.Replicas = deployment.Spec.Replicas
			log.V(1).Info("updating deployment replicacount", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
		} else if foundDeployment.Spec.Template.Spec.Containers[0].Env[0].Value != deployment.Spec.Template.Spec.Containers[0].Env[0].Value { // Check if PodInfo Color was changed
			foundDeployment.Spec.Template.Spec.Containers[0].Env[0].Value = deployment.Spec.Template.Spec.Containers[0].Env[0].Value
			log.V(1).Info("updating deployment podinfo color", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if foundDeployment.Spec.Template.Spec.Containers[0].Env[1].Value != deployment.Spec.Template.Spec.Containers[0].Env[1].Value { // Check if PodInfo Message was changed
			foundDeployment.Spec.Template.Spec.Containers[0].Env[1].Value = deployment.Spec.Template.Spec.Containers[0].Env[1].Value
			log.V(1).Info("updating deployment podinfo message", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if foundDeployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu() != deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu() { // Check if cpuRequest was changed
			foundDeployment.Spec.Template.Spec.Containers[0].Resources.Requests = deployment.Spec.Template.Spec.Containers[0].Resources.Requests
			log.V(1).Info("updating deployment cpuRequest", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if foundDeployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory() != deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory() { // Check if memoryLimit was changed
			foundDeployment.Spec.Template.Spec.Containers[0].Resources.Limits = deployment.Spec.Template.Spec.Containers[0].Resources.Limits
			log.V(1).Info("updating deployment memoryLimit", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Redis logic
	redisDeployment := NewRedisDeploymentForCr(myAppResource)
	redisService := NewRedisServiceForCr(myAppResource)
	redisConfigMap := NewRedisConfigMapForCr(myAppResource)
	if myAppResource.Spec.Redis.Enabled == true {
		err := controllerutil.SetOwnerReference(myAppResource, redisDeployment, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = controllerutil.SetOwnerReference(myAppResource, redisService, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = controllerutil.SetOwnerReference(myAppResource, redisConfigMap, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, redisDeployment)
		if err != nil && errors.IsAlreadyExists(err) != true {
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, redisService)
		if err != nil && errors.IsAlreadyExists(err) != true {
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, redisConfigMap)
		if err != nil && errors.IsAlreadyExists(err) != true {
			return ctrl.Result{}, err
		}

	} else if myAppResource.Spec.Redis.Enabled == false {
		err = r.Delete(ctx, redisDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.Delete(ctx, redisService)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.Delete(ctx, redisConfigMap)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func NewDeploymentForCr(cr *myv1alpha1.MyAppResource) *appsv1.Deployment {
	// Check if CpuRequest is empty
	if cr.Spec.Resources.CpuRequest == "" {
		cr.Spec.Resources.CpuRequest = "100m"
	}
	// Check if MemoryLimit is empty
	if cr.Spec.Resources.MemoryLimit == "" {
		cr.Spec.Resources.MemoryLimit = "64Mi"
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cr.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    cr.Name,
							Image:   cr.Spec.Image.Repository + ":" + cr.Spec.Image.Tag,
							Command: []string{"./podinfo", "--cache-server=http://redis:6379"},
							Env: []apiv1.EnvVar{
								{
									Name:  "PODINFO_UI_COLOR",
									Value: cr.Spec.Ui.Color,
								},
								{
									Name:  "PODINFO_UI_MESSAGE",
									Value: cr.Spec.Ui.Message,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: map[apiv1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse(cr.Spec.Resources.CpuRequest),
									"memory": resource.MustParse(cr.Spec.Resources.CpuRequest),
								},
								Limits: map[apiv1.ResourceName]resource.Quantity{
									"cpu":    resource.MustParse(cr.Spec.Resources.MemoryLimit),
									"memory": resource.MustParse(cr.Spec.Resources.MemoryLimit),
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewRedisDeploymentForCr(cr *myv1alpha1.MyAppResource) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "redis",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "redis",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "redis",
							Image:   "redis:7.0.7",
							Command: []string{"redis-server", "/redis-master/redis.conf"},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "redis",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 6379,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/redis",
								},
								{
									Name:      "config",
									MountPath: "/redis-master",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "data",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "redis",
									},
									DefaultMode: int32Ptr(420),
									Items: []apiv1.KeyToPath{
										{
											Key:  "redis.conf",
											Path: "redis.conf",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewRedisServiceForCr(cr *myv1alpha1.MyAppResource) *apiv1.Service {
	return &apiv1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis",
			Namespace: cr.Namespace,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Name: "redis", Port: 6379, Protocol: "TCP", TargetPort: intstr.FromString("redis")},
			},
			Selector: map[string]string{"app": "redis"},
		},
	}
}

func NewRedisConfigMapForCr(cr *myv1alpha1.MyAppResource) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String(), Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"redis.conf": "maxmemory 64mb\n" +
				"maxmemory-policy allkeys-lru\n" +
				"save \"\"\n" +
				"appendonly no",
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myv1alpha1.MyAppResource{}).
		Owns(&appsv1.Deployment{}).
		Owns(&apiv1.Service{}).
		Owns(&apiv1.ConfigMap{}).
		Complete(r)
}

func int32Ptr(i int32) *int32 { return &i }
