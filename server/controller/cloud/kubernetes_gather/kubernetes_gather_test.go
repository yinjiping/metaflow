package kubernetes_gather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"server/controller/common"
	"server/controller/db/mysql"
	"server/controller/genesis"
	"server/controller/genesis/config"
	genesismodel "server/controller/genesis/model"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKubernetes(t *testing.T) {
	Convey("TestKubernetes", t, func() {
		k8sConfig := mysql.SubDomain{
			Name:        "test_k8s",
			DisplayName: "test_k8s",
			ClusterID:   "d-01LMvvfQPZ",
			Config:      fmt.Sprintf(`{"port_name_regex": "","pod_net_ipv4_cidr_max_mask": %v,"pod_net_ipv6_cidr_max_mask": %v,"region_uuid": "%s","vpc_uuid": ""}`, K8S_POD_IPV4_NETMASK, K8S_POD_IPV6_NETMASK, common.DEFAULT_REGION),
		}

		k8s := NewKubernetesGather(nil, &k8sConfig)
		type KResource struct {
			Pod        []string `json:"*v1.Pod"`
			Info       []string `json:"*version.Info"`
			Node       []string `json:"*v1.Node"`
			Ingress    []string `json:"*v1beta1.Ingress"`
			Service    []string `json:"*v1.Service"`
			ConfigMap  []string `json:"*v1.ConfigMap"`
			DaemonSet  []string `json:"*v1.DaemonSet"`
			Namespace  []string `json:"*v1.Namespace"`
			Deployment []string `json:"*v1.Deployment"`
			ReplicaSet []string `json:"*v1.ReplicaSet"`
		}

		type KDataResp struct {
			ClusterID string    `json:"cluster_id"`
			ErrorMSG  string    `json:"error_msg"`
			SyncedAt  string    `json:"synced_at"`
			Resources KResource `json:"resources"`
		}

		kJsonData, _ := ioutil.ReadFile("./testfiles/kubernetes-info.json")
		var kData KDataResp
		json.Unmarshal(kJsonData, &kData)
		k8sInfo := map[string][]string{}
		k8sInfo["*v1.Pod"] = kData.Resources.Pod
		k8sInfo["*v1.Node"] = kData.Resources.Node
		k8sInfo["*version.Info"] = kData.Resources.Info
		k8sInfo["*v1beta1.Ingress"] = kData.Resources.Ingress
		k8sInfo["*v1.Service"] = kData.Resources.Service
		k8sInfo["*v1.ConfigMap"] = kData.Resources.ConfigMap
		k8sInfo["*v1.DaemonSet"] = kData.Resources.DaemonSet
		k8sInfo["*v1.Namespace"] = kData.Resources.Namespace
		k8sInfo["*v1.Deployment"] = kData.Resources.Deployment
		k8sInfo["*v1.ReplicaSet"] = kData.Resources.ReplicaSet
		k8sInfoPatch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(k8s), "getKubernetesInfo", func(_ *KubernetesGather) (map[string][]string, error) {
			return k8sInfo, nil
		})
		defer k8sInfoPatch.Reset()

		g := genesis.NewGenesis(config.GenesisConfig{})
		type VDataResp struct {
			Desc   string                           `json:"DESCRIPTION"`
			Status string                           `json:"OPT_STATUS"`
			Type   string                           `json:"TYPE"`
			Data   []genesismodel.GenesisVinterface `json:"DATA"`
		}
		vJsonData, _ := ioutil.ReadFile("./testfiles/vinterfaces.json")
		var vData VDataResp
		json.Unmarshal(vJsonData, &vData)
		vinterfacesInfoPatch := gomonkey.ApplyMethod(reflect.TypeOf(g), "GetVinterfacesData", func(_ *genesis.Genesis) []genesismodel.GenesisVinterface {
			return vData.Data
		})
		defer vinterfacesInfoPatch.Reset()

		k8sGatherData, _ := k8s.GetKubernetesGatherData()
		Convey("k8sGatherResource number should be equal", func() {
			So(len(k8sGatherData.PodNodes), ShouldEqual, 2)
			So(len(k8sGatherData.PodNamespaces), ShouldEqual, 7)
			So(len(k8sGatherData.PodGroups), ShouldEqual, 11)
			So(len(k8sGatherData.PodReplicaSets), ShouldEqual, 4)
			So(len(k8sGatherData.PodServices), ShouldEqual, 4)
			So(len(k8sGatherData.PodServicePorts), ShouldEqual, 6)
			So(len(k8sGatherData.PodGroupPorts), ShouldEqual, 6)
			So(len(k8sGatherData.PodIngresses), ShouldEqual, 1)
			So(len(k8sGatherData.PodIngressRules), ShouldEqual, 2)
			So(len(k8sGatherData.PodIngressRuleBackends), ShouldEqual, 2)
			So(len(k8sGatherData.Pods), ShouldEqual, 15)
			So(len(k8sGatherData.PodServiceSubnets), ShouldEqual, 1)
			So(len(k8sGatherData.PodNodeSubnets), ShouldEqual, 1)
			So(len(k8sGatherData.PodSubnets), ShouldEqual, 1)
			So(len(k8sGatherData.PodServiceVInterfaces), ShouldEqual, 4)
			So(len(k8sGatherData.PodNodeVInterfaces), ShouldEqual, 2)
			So(len(k8sGatherData.PodVInterfaces), ShouldEqual, 9)
			So(len(k8sGatherData.PodServiceIPs), ShouldEqual, 4)
			So(len(k8sGatherData.PodNodeIPs), ShouldEqual, 2)
			So(len(k8sGatherData.PodIPs), ShouldEqual, 9)
		})
	})
}
