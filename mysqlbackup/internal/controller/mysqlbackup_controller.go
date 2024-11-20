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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "k8s.io/api/batch/v1"
	webappv1 "xingzai.cn/mysqlbackup/api/v1"
)

// MysqlbackupReconciler reconciles a Mysqlbackup object
type MysqlbackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.xingzai.cn,resources=mysqlbackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.xingzai.cn,resources=mysqlbackups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.xingzai.cn,resources=mysqlbackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mysqlbackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MysqlbackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// TODO(user): your logic here
	job := &webappv1.Mysqlbackup{}
	err := r.Get(ctx, req.NamespacedName, job)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("mysqlbackup resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get mysqlbackup")
		return ctrl.Result{}, err
	}

	//查找同名cronjob是否已经创建
	cronjob := &batchv1.CronJob{}
	if err := r.Get(ctx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, cronjob); err != nil {
		if apierrors.IsNotFound(err) {
			j, err := r.cronjobForMysqlbackup(job)
			if err != nil {
				log.Error(err, "Failed to define new cronjob resource for Mysqlbackup")
				return ctrl.Result{}, err
			}
			//排错使用，检查cronjob生产的是否正确
			log.Info(fmt.Sprintf("%v", j))
			if err = r.Create(ctx, j); err != nil {
				log.Error(err, "create mysql cronjob failed")
				return ctrl.Result{}, err
			} else {

			}
		}
	}

	return ctrl.Result{}, nil
}

// 配置自定义的 cronjob yaml
func (r *MysqlbackupReconciler) cronjobForMysqlbackup(mysqlbackup *webappv1.Mysqlbackup) (*batchv1.CronJob, error) {
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlbackup.Name,
			Namespace: mysqlbackup.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: mysqlbackup.Spec.Cronjob,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            mysqlbackup.Name,
									Image:           "harbor.hzxingzai.cn/tools/mysqldump",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Command:         []string{"/bin/sh", "-c", "/opt/mysqldump.sh"},
									Env: []corev1.EnvVar{
										{
											Name:  "MYSQL_SVC_NAME",
											Value: mysqlbackup.Spec.Mysqlsvcname,
										},
										{
											Name:  "BACKUP_USER",
											Value: mysqlbackup.Spec.Backupuser,
										},
										{
											Name:  "BACKUP_PASSWORD",
											Value: mysqlbackup.Spec.Backuppassword,
										},
										{
											Name:  "BACKUP_DB",
											Value: mysqlbackup.Spec.Backupdb,
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "backup",      // 对应的卷名称
											MountPath: "/opt/backup", // 挂载路径
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Volumes: []corev1.Volume{
								{
									Name: "backup", // 定义 NFS 卷的名称
									VolumeSource: corev1.VolumeSource{
										// 使用 NFS 作为卷源
										NFS: &corev1.NFSVolumeSource{
											Server:   "10.10.70.3",             // NFS 服务器的地址
											Path:     "/datadisk/backup/mysql", // NFS 共享路径
											ReadOnly: false,                    // 可选：设置为 true 表示只读
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
	// Set the ownerRef for the Cronjob
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(mysqlbackup, job, r.Scheme); err != nil {
		return nil, err
	}
	return job, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlbackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Mysqlbackup{}).
		Owns(&batchv1.CronJob{}).
		Named("mysqlbackup").
		Complete(r)
}
