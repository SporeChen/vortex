package pod

import (
	"github.com/linkernetworks/vortex/src/config"
	"github.com/linkernetworks/vortex/src/entity"
	"github.com/linkernetworks/vortex/src/serviceprovider"
	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type PodTestSuite struct {
	suite.Suite
	sp *serviceprovider.Container
}

func (suite *PodTestSuite) SetupSuite() {
	cf := config.MustRead("../../config/testing.json")
	suite.sp = serviceprovider.NewForTesting(cf)
}

func (suite *PodTestSuite) TearDownSuite() {
}

func TestPodSuite(t *testing.T) {
	suite.Run(t, new(PodTestSuite))
}

func (suite *PodTestSuite) TestCheckPodParameter() {
	volumeName := namesgenerator.GetRandomName(0)
	pod := &entity.Pod{
		ID: bson.NewObjectId(),
		Volumes: []entity.PodVolume{
			{Name: volumeName},
		},
	}

	session := suite.sp.Mongo.NewSession()
	defer session.Close()

	volume := entity.Volume{
		ID:   bson.NewObjectId(),
		Name: volumeName,
	}

	session.Insert(entity.VolumeCollectionName, volume)
	defer session.Remove(entity.VolumeCollectionName, "name", volume.Name)

	err := CheckPodParameter(suite.sp, pod)
	suite.NoError(err)
}

func (suite *PodTestSuite) TestCheckPodParameterFail() {
	volumeName := namesgenerator.GetRandomName(0)
	pod := &entity.Pod{
		ID: bson.NewObjectId(),
		Volumes: []entity.PodVolume{
			{Name: volumeName},
		},
	}

	err := CheckPodParameter(suite.sp, pod)
	suite.Error(err)
}

func (suite *PodTestSuite) TestGenerateVolume() {
	volumeName := namesgenerator.GetRandomName(0)
	pod := &entity.Pod{
		ID: bson.NewObjectId(),
		Volumes: []entity.PodVolume{
			{Name: volumeName},
		},
	}

	session := suite.sp.Mongo.NewSession()
	defer session.Close()

	volume := entity.Volume{
		ID:   bson.NewObjectId(),
		Name: volumeName,
	}
	session.Insert(entity.VolumeCollectionName, volume)
	defer session.Remove(entity.VolumeCollectionName, "name", volume.Name)

	volumes, volumeMounts, err := generateVolume(pod, session)
	suite.NotNil(volumes)
	suite.NotNil(volumeMounts)
	suite.NoError(err)
}

func (suite *PodTestSuite) TestGenerateVolumeFail() {
	volumeName := namesgenerator.GetRandomName(0)
	pod := &entity.Pod{
		ID: bson.NewObjectId(),
		Volumes: []entity.PodVolume{
			{Name: volumeName},
		},
	}

	session := suite.sp.Mongo.NewSession()
	defer session.Close()
	volumes, volumeMounts, err := generateVolume(pod, session)
	suite.Nil(volumes)
	suite.Nil(volumeMounts)
	suite.Error(err)
}

func (suite *PodTestSuite) TestCreatePod() {
	containers := []entity.Container{
		{
			Name:    namesgenerator.GetRandomName(0),
			Image:   "busybox",
			Command: []string{"sleep", "3600"},
		},
	}

	podName := namesgenerator.GetRandomName(0)
	pod := &entity.Pod{
		ID:         bson.NewObjectId(),
		Name:       podName,
		Containers: containers,
	}

	err := CreatePod(suite.sp, pod)
	suite.NoError(err)

	err = DeletePod(suite.sp, podName)
	suite.NoError(err)
}

func (suite *PodTestSuite) TestCreatePodFail() {
	containers := []entity.Container{
		{
			Name:    namesgenerator.GetRandomName(0),
			Image:   "busybox",
			Command: []string{"sleep", "3600"},
		},
	}

	podName := namesgenerator.GetRandomName(0)
	pod := &entity.Pod{
		ID:         bson.NewObjectId(),
		Name:       podName,
		Containers: containers,
		Volumes: []entity.PodVolume{
			{Name: namesgenerator.GetRandomName(0)},
		},
	}

	err := CreatePod(suite.sp, pod)
	suite.Error(err)
}