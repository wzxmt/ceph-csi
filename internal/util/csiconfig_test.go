/*
Copyright 2019 ceph-csi authors.

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

package util

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	cephcsi "github.com/ceph/ceph-csi/api/deploy/kubernetes"

	"github.com/stretchr/testify/require"
)

var (
	basePath     = "./test_artifacts"
	csiClusters  = "csi-clusters.json"
	pathToConfig = basePath + "/" + csiClusters
	clusterID1   = "test1"
	clusterID2   = "test2"
)

func cleanupTestData() {
	os.RemoveAll(basePath)
}

func TestCSIConfig(t *testing.T) {
	t.Parallel()
	var err error
	var data string
	var content string

	defer cleanupTestData()

	err = os.MkdirAll(basePath, 0o700)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should fail as clusterid file is missing
	_, err = Mons(pathToConfig, clusterID1)
	if err == nil {
		t.Errorf("Failed: expected error due to missing config")
	}

	data = ""
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should fail as file is empty
	content, err = Mons(pathToConfig, clusterID1)
	if err == nil {
		t.Errorf("Failed: want (%s), got (%s)", data, content)
	}

	data = "[{\"clusterIDBad\":\"" + clusterID2 + "\",\"monitors\":[\"mon1\",\"mon2\",\"mon3\"]}]"
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should fail as clusterID data is malformed
	content, err = Mons(pathToConfig, clusterID2)
	if err == nil {
		t.Errorf("Failed: want (%s), got (%s)", data, content)
	}

	data = "[{\"clusterID\":\"" + clusterID2 + "\",\"monitorsBad\":[\"mon1\",\"mon2\",\"mon3\"]}]"
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should fail as monitors key is incorrect/missing
	content, err = Mons(pathToConfig, clusterID2)
	if err == nil {
		t.Errorf("Failed: want (%s), got (%s)", data, content)
	}

	data = "[{\"clusterID\":\"" + clusterID2 + "\",\"monitors\":[\"mon1\",2,\"mon3\"]}]"
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should fail as monitor data is malformed
	content, err = Mons(pathToConfig, clusterID2)
	if err == nil {
		t.Errorf("Failed: want (%s), got (%s)", data, content)
	}

	data = "[{\"clusterID\":\"" + clusterID2 + "\",\"monitors\":[\"mon1\",\"mon2\",\"mon3\"]}]"
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should fail as clusterID is not present in config
	content, err = Mons(pathToConfig, clusterID1)
	if err == nil {
		t.Errorf("Failed: want (%s), got (%s)", data, content)
	}

	// TEST: Should pass as clusterID is present in config
	content, err = Mons(pathToConfig, clusterID2)
	if err != nil || content != "mon1,mon2,mon3" {
		t.Errorf("Failed: want (%s), got (%s) (%v)", "mon1,mon2,mon3", content, err)
	}

	data = "[{\"clusterID\":\"" + clusterID2 + "\",\"monitors\":[\"mon1\",\"mon2\",\"mon3\"]}," +
		"{\"clusterID\":\"" + clusterID1 + "\",\"monitors\":[\"mon4\",\"mon5\",\"mon6\"]}]"
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}

	// TEST: Should pass as clusterID is present in config
	content, err = Mons(pathToConfig, clusterID1)
	if err != nil || content != "mon4,mon5,mon6" {
		t.Errorf("Failed: want (%s), got (%s) (%v)", "mon4,mon5,mon6", content, err)
	}

	data = "[{\"clusterID\":\"" + clusterID2 + "\",\"monitors\":[\"mon1\",\"mon2\",\"mon3\"]}," +
		"{\"clusterID\":\"" + clusterID1 + "\",\"monitors\":[\"mon4\",\"mon5\",\"mon6\"]}]"
	err = os.WriteFile(basePath+"/"+csiClusters, []byte(data), 0o600)
	if err != nil {
		t.Errorf("Test setup error %s", err)
	}
}

func TestGetRBDNetNamespaceFilePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		clusterID string
		want      string
	}{
		{
			name:      "get RBD NetNamespaceFilePath for cluster-1",
			clusterID: "cluster-1",
			want:      "/var/lib/kubelet/plugins/rbd.ceph.csi.com/cluster1-net",
		},
		{
			name:      "get RBD NetNamespaceFilePath for cluster-2",
			clusterID: "cluster-2",
			want:      "/var/lib/kubelet/plugins/rbd.ceph.csi.com/cluster2-net",
		},
		{
			name:      "when RBD NetNamespaceFilePath is empty",
			clusterID: "cluster-3",
			want:      "",
		},
	}

	csiConfig := []cephcsi.ClusterInfo{
		{
			ClusterID: "cluster-1",
			Monitors:  []string{"ip-1", "ip-2"},
			RBD: cephcsi.RBD{
				NetNamespaceFilePath: "/var/lib/kubelet/plugins/rbd.ceph.csi.com/cluster1-net",
			},
		},
		{
			ClusterID: "cluster-2",
			Monitors:  []string{"ip-3", "ip-4"},
			RBD: cephcsi.RBD{
				NetNamespaceFilePath: "/var/lib/kubelet/plugins/rbd.ceph.csi.com/cluster2-net",
			},
		},
		{
			ClusterID: "cluster-3",
			Monitors:  []string{"ip-5", "ip-6"},
		},
	}
	csiConfigFileContent, err := json.Marshal(csiConfig)
	if err != nil {
		t.Errorf("failed to marshal csi config info %v", err)
	}
	tmpConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetRBDNetNamespaceFilePath(tmpConfPath, tt.clusterID)
			if err != nil {
				t.Errorf("GetRBDNetNamespaceFilePath() error = %v", err)

				return
			}
			if got != tt.want {
				t.Errorf("GetRBDNetNamespaceFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCephFSNetNamespaceFilePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		clusterID string
		want      string
	}{
		{
			name:      "get cephFS specific NetNamespaceFilePath for cluster-1",
			clusterID: "cluster-1",
			want:      "/var/lib/kubelet/plugins/cephfs.ceph.csi.com/cluster1-net",
		},
		{
			name:      "get cephFS specific NetNamespaceFilePath for cluster-2",
			clusterID: "cluster-2",
			want:      "/var/lib/kubelet/plugins/cephfs.ceph.csi.com/cluster2-net",
		},
		{
			name:      "when cephFS specific NetNamespaceFilePath is empty",
			clusterID: "cluster-3",
			want:      "",
		},
	}

	csiConfig := []cephcsi.ClusterInfo{
		{
			ClusterID: "cluster-1",
			Monitors:  []string{"ip-1", "ip-2"},
			CephFS: cephcsi.CephFS{
				NetNamespaceFilePath: "/var/lib/kubelet/plugins/cephfs.ceph.csi.com/cluster1-net",
			},
		},
		{
			ClusterID: "cluster-2",
			Monitors:  []string{"ip-3", "ip-4"},
			CephFS: cephcsi.CephFS{
				NetNamespaceFilePath: "/var/lib/kubelet/plugins/cephfs.ceph.csi.com/cluster2-net",
			},
		},
		{
			ClusterID: "cluster-3",
			Monitors:  []string{"ip-5", "ip-6"},
		},
	}
	csiConfigFileContent, err := json.Marshal(csiConfig)
	if err != nil {
		t.Errorf("failed to marshal csi config info %v", err)
	}
	tmpConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetCephFSNetNamespaceFilePath(tmpConfPath, tt.clusterID)
			if err != nil {
				t.Errorf("GetCephFSNetNamespaceFilePath() error = %v", err)

				return
			}
			if got != tt.want {
				t.Errorf("GetCephFSNetNamespaceFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNFSNetNamespaceFilePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		clusterID string
		want      string
	}{
		{
			name:      "get NFS specific NetNamespaceFilePath for cluster-1",
			clusterID: "cluster-1",
			want:      "/var/lib/kubelet/plugins/nfs.ceph.csi.com/cluster1-net",
		},
		{
			name:      "get NFS specific NetNamespaceFilePath for cluster-2",
			clusterID: "cluster-2",
			want:      "/var/lib/kubelet/plugins/nfs.ceph.csi.com/cluster2-net",
		},
		{
			name:      "when NFS specific NetNamespaceFilePath is empty",
			clusterID: "cluster-3",
			want:      "",
		},
	}

	csiConfig := []cephcsi.ClusterInfo{
		{
			ClusterID: "cluster-1",
			Monitors:  []string{"ip-1", "ip-2"},
			NFS: cephcsi.NFS{
				NetNamespaceFilePath: "/var/lib/kubelet/plugins/nfs.ceph.csi.com/cluster1-net",
			},
		},
		{
			ClusterID: "cluster-2",
			Monitors:  []string{"ip-3", "ip-4"},
			NFS: cephcsi.NFS{
				NetNamespaceFilePath: "/var/lib/kubelet/plugins/nfs.ceph.csi.com/cluster2-net",
			},
		},
		{
			ClusterID: "cluster-3",
			Monitors:  []string{"ip-5", "ip-6"},
		},
	}
	csiConfigFileContent, err := json.Marshal(csiConfig)
	if err != nil {
		t.Errorf("failed to marshal csi config info %v", err)
	}
	tmpConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetNFSNetNamespaceFilePath(tmpConfPath, tt.clusterID)
			if err != nil {
				t.Errorf("GetNFSNetNamespaceFilePath() error = %v", err)

				return
			}
			if got != tt.want {
				t.Errorf("GetNFSNetNamespaceFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReadAffinityOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		clusterID string
		want      struct {
			enabled bool
			labels  string
		}
	}{
		{
			name:      "ReadAffinity enabled set to true for cluster-1",
			clusterID: "cluster-1",
			want: struct {
				enabled bool
				labels  string
			}{true, "topology.kubernetes.io/region,topology.kubernetes.io/zone,topology.io/rack"},
		},
		{
			name:      "ReadAffinity enabled set to true for cluster-2",
			clusterID: "cluster-2",
			want: struct {
				enabled bool
				labels  string
			}{true, "topology.kubernetes.io/region"},
		},
		{
			name:      "ReadAffinity enabled set to false for cluster-3",
			clusterID: "cluster-3",
			want: struct {
				enabled bool
				labels  string
			}{false, ""},
		},
		{
			name:      "ReadAffinity option not set in cluster-4",
			clusterID: "cluster-4",
			want: struct {
				enabled bool
				labels  string
			}{false, ""},
		},
	}

	csiConfig := []cephcsi.ClusterInfo{
		{
			ClusterID: "cluster-1",
			ReadAffinity: cephcsi.ReadAffinity{
				Enabled: true,
				CrushLocationLabels: []string{
					"topology.kubernetes.io/region",
					"topology.kubernetes.io/zone",
					"topology.io/rack",
				},
			},
		},
		{
			ClusterID: "cluster-2",
			ReadAffinity: cephcsi.ReadAffinity{
				Enabled: true,
				CrushLocationLabels: []string{
					"topology.kubernetes.io/region",
				},
			},
		},
		{
			ClusterID: "cluster-3",
			ReadAffinity: cephcsi.ReadAffinity{
				Enabled: false,
				CrushLocationLabels: []string{
					"topology.io/rack",
				},
			},
		},
		{
			ClusterID: "cluster-4",
		},
	}
	csiConfigFileContent, err := json.Marshal(csiConfig)
	if err != nil {
		t.Errorf("failed to marshal csi config info %v", err)
	}
	tmpConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			enabled, labels, err := GetCrushLocationLabels(tmpConfPath, tt.clusterID)
			if err != nil {
				t.Errorf("GetCrushLocationLabels() error = %v", err)

				return
			}
			if enabled != tt.want.enabled || labels != tt.want.labels {
				t.Errorf("GetCrushLocationLabels() = {%v %v} want %v", enabled, labels, tt.want)
			}
		})
	}
}

func TestGetCephFSMountOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		clusterID            string
		wantKernelMntOptions string
		wantFuseMntOptions   string
	}{
		{
			name:                 "cluster-1 with non-empty mount options",
			clusterID:            "cluster-1",
			wantKernelMntOptions: "crc",
			wantFuseMntOptions:   "ro",
		},
		{
			name:                 "cluster-2 with empty mount options",
			clusterID:            "cluster-2",
			wantKernelMntOptions: "",
			wantFuseMntOptions:   "",
		},
		{
			name:                 "cluster-3 with no mount options",
			clusterID:            "cluster-3",
			wantKernelMntOptions: "",
			wantFuseMntOptions:   "",
		},
	}

	csiConfig := []cephcsi.ClusterInfo{
		{
			ClusterID: "cluster-1",
			CephFS: cephcsi.CephFS{
				KernelMountOptions: "crc",
				FuseMountOptions:   "ro",
			},
		},
		{
			ClusterID: "cluster-2",
			CephFS: cephcsi.CephFS{
				KernelMountOptions: "",
				FuseMountOptions:   "",
			},
		},
		{
			ClusterID: "cluster-3",
			CephFS:    cephcsi.CephFS{},
		},
	}
	csiConfigFileContent, err := json.Marshal(csiConfig)
	if err != nil {
		t.Errorf("failed to marshal csi config info %v", err)
	}
	tmpConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			kernelMntOptions, fuseMntOptions, err := GetCephFSMountOptions(tmpConfPath, tt.clusterID)
			if err != nil {
				t.Errorf("GetCephFSMountOptions() error = %v", err)
			}
			if kernelMntOptions != tt.wantKernelMntOptions || fuseMntOptions != tt.wantFuseMntOptions {
				t.Errorf("GetCephFSMountOptions() = (%v, %v), want (%v, %v)",
					kernelMntOptions, fuseMntOptions, tt.wantKernelMntOptions, tt.wantFuseMntOptions,
				)
			}
		})
	}
}

func TestGetRBDMirrorDaemonCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		clusterID string
		want      int
	}{
		{
			name:      "get rbd mirror daemon count for cluster-1",
			clusterID: "cluster-1",
			want:      2,
		},
		{
			name:      "get rbd mirror daemon count for cluster-2",
			clusterID: "cluster-2",
			want:      4,
		},
		{
			name:      "when rbd mirror daemon count is empty",
			clusterID: "cluster-3",
			want:      1, // default mirror daemon count
		},
	}

	csiConfig := []cephcsi.ClusterInfo{
		{
			ClusterID: "cluster-1",
			Monitors:  []string{"ip-1", "ip-2"},
			RBD: cephcsi.RBD{
				MirrorDaemonCount: 2,
			},
		},
		{
			ClusterID: "cluster-2",
			Monitors:  []string{"ip-3", "ip-4"},
			RBD: cephcsi.RBD{
				MirrorDaemonCount: 4,
			},
		},
		{
			ClusterID: "cluster-3",
			Monitors:  []string{"ip-5", "ip-6"},
		},
	}
	csiConfigFileContent, err := json.Marshal(csiConfig)
	if err != nil {
		t.Errorf("failed to marshal csi config info %v", err)
	}
	tmpConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got int
			got, err = GetRBDMirrorDaemonCount(tmpConfPath, tt.clusterID)
			if err != nil {
				t.Errorf("GetRBDMirrorDaemonCount() error = %v", err)

				return
			}
			if got != tt.want {
				t.Errorf("GetRBDMirrorDaemonCount() = %v, want %v", got, tt.want)
			}
		})
	}

	// when mirrorDaemonCount is set as string
	csiConfigFileContent = bytes.Replace(
		csiConfigFileContent,
		[]byte(`"mirrorDaemonCount":2`),
		[]byte(`"mirrorDaemonCount":"2"`),
		1)
	tmpCSIConfPath := t.TempDir() + "/ceph-csi.json"
	err = os.WriteFile(tmpCSIConfPath, csiConfigFileContent, 0o600)
	if err != nil {
		t.Errorf("failed to write %s file content: %v", CsiConfigFile, err)
	}
	_, err = GetRBDMirrorDaemonCount(tmpCSIConfPath, "test")
	require.Error(t, err)
}
