package service

import (
	"path"

	xctx "golang.org/x/net/context"

	"github.com/thecodeteam/gocsi/csi"
)

func (s *service) GetVolumeName(
	ctx xctx.Context,
	id string) (string, error) {

	i, v := s.findVol("id", id)
	if i < 0 {
		return "", nil
	}
	return v.Attributes["name"], nil
}

func (s *service) GetVolumeInfo(
	ctx xctx.Context,
	name string) (*csi.VolumeInfo, error) {

	i, v := s.findVol("name", name)
	if i < 0 {
		return nil, nil
	}
	return &v, nil
}

func (s *service) IsControllerPublished(
	ctx xctx.Context,
	id, nodeID string) (map[string]string, error) {

	_, v := s.findVol("id", id)
	if _, ok := v.Attributes[path.Join(nodeID, "dev")]; ok {
		return map[string]string{"device": "/dev/mock"}, nil
	}
	return nil, nil
}

func (s *service) IsNodePublished(
	ctx xctx.Context,
	id string,
	pubInfo map[string]string,
	targetPath string) (bool, error) {

	_, v := s.findVol("id", id)
	if _, ok := v.Attributes[path.Join(s.nodeID, targetPath)]; ok {
		return true, nil
	}
	return false, nil
}
