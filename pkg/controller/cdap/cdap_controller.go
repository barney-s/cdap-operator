package cdap

import (
	"context"

	iov1alpha1 "io.cdap/cdap-operator/pkg/apis/io/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("cdap.controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CDAP Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCDAP{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cdap-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CDAP
	err = c.Watch(&source.Kind{Type: &iov1alpha1.CDAP{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSet and requeue the owner CDAP
	err = c.Watch(&source.Kind{Type: &appv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &iov1alpha1.CDAP{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCDAP{}

// ReconcileCDAP reconciles a CDAP object
type ReconcileCDAP struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a CDAP object and makes changes based on the state read
// and what is in the CDAP.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCDAP) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CDAP")

	// Fetch the CDAP instance
	instance := &iov1alpha1.CDAP{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for _, service := range instance.Spec.Services {
		serviceStatefulSet := createStateSetForService(instance.Name, request.Namespace, instance.Spec.Image, &service)
		reqLogger.Info("Service", "type", service.Type, "StatefulSet", serviceStatefulSet)
	}

	// Define a new StatefulSet object
	statefulSet := newStatefulSetForCR(instance)

	// Set CDAP instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, statefulSet, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this StatefulSet already exists
	found := &appv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: statefulSet.Name, Namespace: statefulSet.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Deploying new StatefulSet", "Namespace", statefulSet.Namespace, "Name", statefulSet.Name)
		err = r.client.Create(context.TODO(), statefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}

		// StatefulSet created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// StatefulSet already exists - don't requeue
	reqLogger.Info("Skip reconcile: StatefulSet already exists", "Namespace", found.Namespace, "Name", found.Name)
	return reconcile.Result{}, nil
}

func createStateSetForService(instanceName string, namespace string, image string, service *iov1alpha1.CDAPService) *appv1.StatefulSet {
	name := instanceName + "-" + string(service.Type)
	labels := map[string]string{
		"app": name,
	}
	log.Info("Instance name", "name", name)
	return &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: service.Instances,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    name,
							Image:   image,
							Command: []string{"sh", "-c", "/opt/cdap/sandbox/bin/cdap sandbox start --foreground"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/opt/cdap/sandbox-5.1.0/data",
									SubPath:   "cdap",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
}

func newStatefulSetForCR(cr *iov1alpha1.CDAP) *appv1.StatefulSet {
	labels := map[string]string{
		"app": cr.Name,
	}
	replicas := int32(1)

	return &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    cr.Name,
							Image:   cr.Spec.Image,
							Command: []string{"sh", "-c", "/opt/cdap/sandbox/bin/cdap sandbox start --foreground"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 11015,
									Name:          cr.Name,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/opt/cdap/sandbox-5.1.0/data",
									SubPath:   "cdap",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
}
