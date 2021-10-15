package javaautoinstrumentationoperator

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func getOtJarsVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      opentelemetryJarVolumeName,
		MountPath: opentelemetryJarMountPath,
		ReadOnly:  false,
	}
}

func getOtJarsVolume() corev1.Volume {
	return corev1.Volume{
		Name: opentelemetryJarVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func mergePodSpec(originalPodSpec *corev1.PodSpec, serviceName string, exporter string,
	collectorHost string) corev1.PodSpec {

	originalContainer := originalPodSpec.Containers[0] // TODO

	existingJavaOptions, exists := getJavaOptions(originalPodSpec.Containers)
	if !exists {
		existingJavaOptions = ""
	}
	var envVars = copyExistingEnvVarsWithoutJavaOptions(originalContainer.Env)
	envVars = append(envVars, getConfiguration(exporter, serviceName, existingJavaOptions, collectorHost)...)

	var volumes []corev1.Volume
	volumes = append(volumes, originalPodSpec.Volumes...)
	volumes = append(volumes, getOtJarsVolume())

	otJarsVolumeMount := getOtJarsVolumeMount()
	var volumeMounts []corev1.VolumeMount
	volumeMounts = append(volumeMounts, originalContainer.VolumeMounts...)
	volumeMounts = append(volumeMounts, otJarsVolumeMount)

	otJarInitContainerCommand := []string{"/bin/sh", "-c", fmt.Sprintf("cp %s %s", opentelemetryJarPath, opentelemetryJarFinalPath)}
	otJarInitContainerSpec := getOtJarsInitContainer(otJarsVolumeMount, otJarInitContainerCommand)
	otJarInitContainer := append(originalPodSpec.InitContainers, otJarInitContainerSpec)

	return corev1.PodSpec{
		Volumes:        volumes,
		InitContainers: otJarInitContainer,
		Containers: []corev1.Container{
			{
				Name:                     originalContainer.Name,
				Image:                    originalContainer.Image,
				Resources:                originalContainer.Resources,
				SecurityContext:          originalContainer.SecurityContext,
				Env:                      envVars,
				VolumeMounts:             volumeMounts,
				Command:                  originalContainer.Command,
				Args:                     originalContainer.Args,
				WorkingDir:               originalContainer.WorkingDir,
				Ports:                    originalContainer.Ports,
				EnvFrom:                  originalContainer.EnvFrom,
				VolumeDevices:            originalContainer.VolumeDevices,
				LivenessProbe:            originalContainer.LivenessProbe,
				ReadinessProbe:           originalContainer.ReadinessProbe,
				StartupProbe:             originalContainer.StartupProbe,
				Lifecycle:                originalContainer.Lifecycle,
				TerminationMessagePath:   originalContainer.TerminationMessagePath,
				TerminationMessagePolicy: originalContainer.TerminationMessagePolicy,
				ImagePullPolicy:          originalContainer.ImagePullPolicy,
				Stdin:                    originalContainer.Stdin,
				StdinOnce:                originalContainer.StdinOnce,
				TTY:                      originalContainer.TTY,
			},
		},
		EphemeralContainers:           originalPodSpec.EphemeralContainers,
		RestartPolicy:                 originalPodSpec.RestartPolicy,
		TerminationGracePeriodSeconds: originalPodSpec.TerminationGracePeriodSeconds,
		ActiveDeadlineSeconds:         originalPodSpec.ActiveDeadlineSeconds,
		DNSPolicy:                     originalPodSpec.DNSPolicy,
		NodeSelector:                  originalPodSpec.NodeSelector,
		ServiceAccountName:            originalPodSpec.ServiceAccountName,
		DeprecatedServiceAccount:      originalPodSpec.DeprecatedServiceAccount,
		AutomountServiceAccountToken:  originalPodSpec.AutomountServiceAccountToken,
		NodeName:                      originalPodSpec.NodeName,
		HostNetwork:                   originalPodSpec.HostNetwork,
		HostPID:                       originalPodSpec.HostPID,
		HostIPC:                       originalPodSpec.HostIPC,
		ShareProcessNamespace:         originalPodSpec.ShareProcessNamespace,
		SecurityContext:               originalPodSpec.SecurityContext,
		ImagePullSecrets:              originalPodSpec.ImagePullSecrets,
		Hostname:                      originalPodSpec.Hostname,
		Subdomain:                     originalPodSpec.Subdomain,
		Affinity:                      originalPodSpec.Affinity,
		SchedulerName:                 originalPodSpec.SchedulerName,
		Tolerations:                   originalPodSpec.Tolerations,
		HostAliases:                   originalPodSpec.HostAliases,
		PriorityClassName:             originalPodSpec.PriorityClassName,
		Priority:                      originalPodSpec.Priority,
		DNSConfig:                     originalPodSpec.DNSConfig,
		ReadinessGates:                originalPodSpec.ReadinessGates,
		RuntimeClassName:              originalPodSpec.RuntimeClassName,
		EnableServiceLinks:            originalPodSpec.EnableServiceLinks,
		PreemptionPolicy:              originalPodSpec.PreemptionPolicy,
		Overhead:                      originalPodSpec.Overhead,
		TopologySpreadConstraints:     originalPodSpec.TopologySpreadConstraints,
	}
}

func getOtJarsInitContainer(volumeMount corev1.VolumeMount, command []string) corev1.Container {
	return corev1.Container{
		Name:         opentelemetryJarContainerName,
		Image:        opentelemetryJarContainerImage,
		VolumeMounts: []corev1.VolumeMount{volumeMount},
		Command:      command,
	}
}
