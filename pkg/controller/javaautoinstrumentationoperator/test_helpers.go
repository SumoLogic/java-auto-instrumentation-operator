package javaautoinstrumentationoperator

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildContainer(name string, value string) *corev1.Container {
	return &corev1.Container{
		Env: []corev1.EnvVar{
			{
				Name:  name,
				Value: value,
			},
		},
	}
}

func buildDeployment(labels map[string]string) *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "auth-service-abc12-xyz3",
			Namespace:         "ns1",
			UID:               "33333",
			CreationTimestamp: metav1.Now(),
			ClusterName:       "cluster1",
			Labels:            labels,
			Annotations: map[string]string{
				"annotation1": "av1",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "ReplicaSet",
					Name: "foo-bar-rs",
					UID:  "1a1658f9-7818-11e9-90f1-02324f7e0d1e",
				},
			},
		},
		Spec: appv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Hostname: "podHost",
					Containers: []corev1.Container{
						{
							Name: "container1",
						},
						{
							Name: "container2",
						},
					},
				},
			},
		},
	}
}

func buildPodSpecForIntegration() corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "my container",
				Image: "cool docker image",
				Env: []corev1.EnvVar{
					{
						Name:  "my env value",
						Value: "something",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "original volume",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/my-bin",
					},
				},
			},
		},
	}
}
