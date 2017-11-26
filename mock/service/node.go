package service

import (
	"path"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	device, ok := req.PublishVolumeInfo["device"]
	if !ok {
		return nil, status.Error(
			codes.InvalidArgument,
			"publish volume info 'device' key required")
	}

	// nodeMntPathKey is the key in the volume's attributes that is set to a
	// mock mount path if the volume has been published by the node
	nodeMntPathKey := path.Join(s.nodeID, req.TargetPath)

	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	// Publish the volume.
	i, v := s.findVolNoLock("id", req.VolumeId)
	v.Attributes[nodeMntPathKey] = device
	s.vols[i] = v

	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	// nodeMntPathKey is the key in the volume's attributes that is set to a
	// mock mount path if the volume has been published by the node
	nodeMntPathKey := path.Join(s.nodeID, req.TargetPath)

	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	// Unpublish the volume.
	i, v := s.findVolNoLock("id", req.VolumeId)
	delete(v.Attributes, nodeMntPathKey)
	s.vols[i] = v

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *service) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return &csi.GetNodeIDResponse{
		NodeId: s.nodeID,
	}, nil
}

func (s *service) NodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest) (
	*csi.NodeProbeResponse, error) {

	return &csi.NodeProbeResponse{}, nil
}

func (s *service) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			&csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}
