package pod

import (
	"fmt"

	"github.com/linkernetworks/mongo"
	"github.com/linkernetworks/vortex/src/entity"
	"github.com/linkernetworks/vortex/src/serviceprovider"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/mgo.v2/bson"
)

const VolumeNamePrefix = "volume-"

func CheckPodParameter(sp *serviceprovider.Container, pod *entity.Pod) error {
	session := sp.Mongo.NewSession()
	defer session.Close()
	//Check the volume
	for _, v := range pod.Volumes {
		count, err := session.Count(entity.VolumeCollectionName, bson.M{"name": v.Name})
		if err != nil {
			return fmt.Errorf("Check the volume name error:%v", err)
		} else if count == 0 {
			return fmt.Errorf("The volume name %s doesn't exist", v.Name)
		}
	}
	return nil
}

func generateVolume(pod *entity.Pod, session *mongo.Session) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	for i, v := range pod.Volumes {
		volume := entity.Volume{}
		if err := session.FindOne(entity.VolumeCollectionName, bson.M{"name": v.Name}, &volume); err != nil {
			return nil, nil, fmt.Errorf("Get the volume object error:%v", err)
		}

		vName := fmt.Sprintf("%s-%d", VolumeNamePrefix, i)

		volumes = append(volumes, corev1.Volume{
			Name: vName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: volume.GenerateMetaName(),
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      vName,
			MountPath: v.MountPath,
		})
	}

	return volumes, volumeMounts, nil
}

func CreatePod(sp *serviceprovider.Container, pod *entity.Pod) error {

	session := sp.Mongo.NewSession()
	defer session.Close()

	volumes, volumeMounts, err := generateVolume(pod, session)
	if err != nil {
		return err
	}

	var containers []corev1.Container
	for _, container := range pod.Containers {
		containers = append(containers, corev1.Container{
			Name:         container.Name,
			Image:        container.Image,
			Command:      container.Command,
			VolumeMounts: volumeMounts,
		})
	}
	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod.Name,
		},
		Spec: corev1.PodSpec{
			Containers: containers,
			Volumes:    volumes,
		},
	}
	_, err = sp.KubeCtl.CreatePod(&p)
	return err
}

func DeletePod(sp *serviceprovider.Container, podName string) error {
	return sp.KubeCtl.DeletePod(podName)
}