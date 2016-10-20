package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/sjpotter/infranetes/pkg/common"
	"github.com/sjpotter/infranetes/pkg/infranetes/provider"

	dockerclient "github.com/docker/engine-api/client"
	dockertypes "github.com/docker/engine-api/types"

	kubeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

type dockerProvider struct {
}

var client *dockerclient.Client

func init() {
	provider.ImageProviders.RegisterProvider("docker", NewDockerProvider)
}

func NewDockerProvider() (provider.ImageProvider, error) {
	var err error
	if client, err = dockerclient.NewClient(dockerclient.DefaultDockerHost, "", nil, nil); err != nil {
		return nil, err
	}

	return &dockerProvider{}, nil
}

func (d *dockerProvider) CreateContainer(req *kubeapi.CreateContainerRequest) (*kubeapi.CreateContainerResponse, error) {
	return nil, errors.New("Not Implemented")
}

func (d *dockerProvider) StartContainer(req *kubeapi.StartContainerRequest) (*kubeapi.StartContainerResponse, error) {
	return nil, errors.New("Not Implemented")
}

func (d *dockerProvider) StopContainer(req *kubeapi.StopContainerRequest) (*kubeapi.StopContainerResponse, error) {
	return nil, errors.New("Not Implemented")
}

func (d *dockerProvider) RemoveContainer(req *kubeapi.RemoveContainerRequest) (*kubeapi.RemoveContainerResponse, error) {
	return nil, errors.New("Not Implemented")
}

func (d *dockerProvider) ListContainers(req *kubeapi.ListContainersRequest) (*kubeapi.ListContainersResponse, error) {
	return nil, errors.New("Not Implemented")
}

func (d *dockerProvider) ContainerStatus(req *kubeapi.ContainerStatusRequest) (*kubeapi.ContainerStatusResponse, error) {
	return nil, errors.New("Not Implemented")
}

func (d *dockerProvider) Exec(sstream kubeapi.RuntimeService_ExecServer) error {
	return errors.New("Not Implemented")
}

func (d *dockerProvider) ListImages(req *kubeapi.ListImagesRequest) (*kubeapi.ListImagesResponse, error) {
	opts := dockertypes.ImageListOptions{}

	filter := req.Filter
	if filter != nil {
		if imgSpec := filter.GetImage(); imgSpec != nil {
			opts.MatchName = imgSpec.GetImage()
		}
	}

	images, err := client.ImageList(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	result := []*kubeapi.Image{}
	for _, i := range images {
		apiImage, err := common.ToRuntimeAPIImage(&i)
		if err != nil {
			// TODO: log an error message?
			continue
		}
		result = append(result, apiImage)
	}

	resp := &kubeapi.ListImagesResponse{
		Images: result,
	}

	return resp, nil
}

func (d *dockerProvider) ImageStatus(req *kubeapi.ImageStatusRequest) (*kubeapi.ImageStatusResponse, error) {
	newreq := &kubeapi.ListImagesRequest{
		Filter: &kubeapi.ImageFilter{
			Image: req.Image,
		},
	}
	listresp, err := d.ListImages(newreq)
	if err != nil {
		return nil, err
	}
	images := listresp.Images
	if len(images) != 1 {
		return nil, fmt.Errorf("ImageStatus returned more than one image: %+v", images)
	}

	resp := &kubeapi.ImageStatusResponse{
		Image: images[0],
	}
	return resp, nil
}

func (d *dockerProvider) PullImage(req *kubeapi.PullImageRequest) (*kubeapi.PullImageResponse, error) {
	pullresp, err := client.ImagePull(context.Background(), req.Image.GetImage(), dockertypes.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("ImagePull Failed (%v)\n", err)
	}

	decoder := json.NewDecoder(pullresp)
	for {
		var msg interface{}
		err := decoder.Decode(&msg)

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Pull Image failed: %v", err)
		}
	}

	pullresp.Close()

	resp := &kubeapi.PullImageResponse{}

	return resp, err
}

func (d *dockerProvider) RemoveImage(req *kubeapi.RemoveImageRequest) (*kubeapi.RemoveImageResponse, error) {
	_, err := client.ImageRemove(context.Background(), req.Image.GetImage(), dockertypes.ImageRemoveOptions{PruneChildren: true})

	resp := &kubeapi.RemoveImageResponse{}

	return resp, err
}
