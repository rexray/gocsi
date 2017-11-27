package service

import (
	"context"
	"path"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) GetVolumeID(
	ctx context.Context,
	name string) (string, error) {

	i, v := s.findVol("name", name)
	if i < 0 {
		return "", nil
	}
	return v.Id, nil
}

func (s *service) GetVolumeInfo(
	ctx context.Context,
	id, name string) (*csi.VolumeInfo, error) {

	var (
		i = -1
		v csi.VolumeInfo
	)
	if id != "" {
		i, v = s.findVol("id", id)
	}
	if i < 0 && name != "" {
		i, v = s.findVol("name", name)
	}

	if i < 0 {
		return nil, nil
	}

	return &v, nil
}

func (s *service) IsControllerPublished(
	ctx context.Context,
	id, nodeID string) (map[string]string, error) {

	_, v := s.findVol("id", id)
	if _, ok := v.Attributes[path.Join(nodeID, "dev")]; ok {
		return map[string]string{"device": "/dev/mock"}, nil
	}
	return nil, nil
}

func (s *service) IsNodePublished(
	ctx context.Context,
	id string,
	pubInfo map[string]string,
	targetPath string) (bool, error) {

	_, v := s.findVol("id", id)
	if _, ok := v.Attributes[path.Join(s.nodeID, targetPath)]; ok {
		return true, nil
	}
	return false, nil
}
