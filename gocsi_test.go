package gocsi_test

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

const (
	mockPkg    = "github.com/codedellemc/gocsi/mock"
	pluginName = "mock"
)

var mockBinPath = os.Getenv("GOCSI_MOCK")

func init() {
	if mockBinPath == "" {
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		mockBinPath = path.Join(path.Dir(exe), "mock")
		out, err := exec.Command(
			"go", "build", "-o", mockBinPath, mockPkg).CombinedOutput()
		if err != nil {
			panic(fmt.Errorf("error: build mock failed: %v\n%v",
				err, string(out)))
		}
	}
	if _, err := os.Stat(mockBinPath); err != nil {
		panic(err)
	}
}

func startMockServer(
	ctx context.Context) (*grpc.ClientConn, func(), error) {

	f, _ := ioutil.TempFile("", "")
	sockFile := f.Name()
	os.RemoveAll(sockFile)
	endpoint := fmt.Sprintf("unix://%s", sockFile)

	cmd := exec.Command(mockBinPath)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("CSI_ENDPOINT=%s", endpoint))
	//cmd.Run()
	stdout, err := cmd.StdoutPipe()
	Ω(err).ShouldNot(HaveOccurred())

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	started := false
	scan := bufio.NewScanner(stdout)
	for scan.Scan() {
		if strings.Contains(scan.Text(), pluginName+".Serve:") {
			started = true
			break
		}
	}

	if !started {
		return nil, nil, cmd.Wait()
	}

	client, err := gocsi.NewGrpcClient(ctx, endpoint, true)
	if err != nil {
		return nil, nil, cmd.Wait()
	}

	stopMock := func() {
		Ω(cmd.Process.Signal(os.Interrupt)).ShouldNot(HaveOccurred())
		Ω(cmd.Wait()).ShouldNot(HaveOccurred())
		os.RemoveAll(sockFile)
	}

	return client, stopMock, nil
}

func newCSIVersion(major, minor, patch uint32) *csi.Version {
	return &csi.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

var mockSupportedVersions = []*csi.Version{
	newCSIVersion(0, 1, 0),
	newCSIVersion(0, 2, 0),
	newCSIVersion(1, 0, 0),
	newCSIVersion(1, 1, 0),
}
