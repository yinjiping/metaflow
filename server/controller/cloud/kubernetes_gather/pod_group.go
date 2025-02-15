package kubernetes_gather

import (
	cloudcommon "server/controller/cloud/common"
	"server/controller/cloud/model"
	"server/controller/common"
	"strings"

	"github.com/bitly/go-simplejson"
	mapset "github.com/deckarep/golang-set"
	uuid "github.com/satori/go.uuid"
)

func (k *KubernetesGather) getPodGroups() (podGroups []model.PodGroup, err error) {
	log.Debug("get podgroups starting")
	podControllers := [3][]string{}
	podControllers[0] = k.k8sInfo["*v1.Deployment"]
	podControllers[1] = k.k8sInfo["*v1.StatefulSet"]
	podControllers[2] = k.k8sInfo["*v1.DaemonSet"]
	for t, podController := range podControllers {
		for _, c := range podController {
			cData, cErr := simplejson.NewJson([]byte(c))
			if cErr != nil {
				err = cErr
				log.Errorf("podgroup initialization simplejson error: (%s)", cErr.Error())
				return
			}
			metaData, ok := cData.CheckGet("metadata")
			if !ok {
				log.Info("podgroup metadata not found")
				continue
			}
			uID := metaData.Get("uid").MustString()
			if uID == "" {
				log.Info("podgroup uid not found")
				continue
			}
			name := metaData.Get("name").MustString()
			if name == "" {
				log.Infof("podgroup (%s) name not found", uID)
				continue
			}
			namespace := metaData.Get("namespace").MustString()
			if namespace == "" {
				log.Infof("podgroup (%s) namespace not found", name)
				continue
			}
			namespaceLcuuid, ok := k.namespaceToLcuuid[namespace]
			if !ok {
				log.Infof("podgroup (%s) namespace id not found", name)
				continue
			}
			serviceType := common.POD_GROUP_STATEFULSET
			label := "statefulset" + namespace + ":" + name
			replicas := cData.Get("spec").Get("replicas").MustInt()
			switch t {
			case 0:
				serviceType = common.POD_GROUP_DEPLOYMENT
				label = "deployment" + namespace + ":" + name
			case 2:
				replicas = 0
				serviceType = common.POD_GROUP_DAEMON_SET
				label = "daemonset" + namespace + ":" + name
			}
			_, ok = k.nsLabelToGroupLcuuids[namespace+label]
			if ok {
				k.nsLabelToGroupLcuuids[namespace+label].Add(uID)
			} else {
				groupIDsSet := mapset.NewSet()
				groupIDsSet.Add(uID)
				k.nsLabelToGroupLcuuids[namespace+label] = groupIDsSet
			}
			mLabels := cData.Get("spec").Get("selector").Get("matchLabels").MustMap()
			for key, v := range mLabels {
				nsLabel := namespace + key + "_" + v.(string)
				_, ok := k.nsLabelToGroupLcuuids[nsLabel]
				if ok {
					k.nsLabelToGroupLcuuids[nsLabel].Add(uID)
				} else {
					nsGroupIDsSet := mapset.NewSet()
					nsGroupIDsSet.Add(uID)
					k.nsLabelToGroupLcuuids[nsLabel] = nsGroupIDsSet
				}
			}
			containersPorts := cData.Get("spec").Get("template").Get("spec").Get("containers")
			for i := range containersPorts.MustArray() {
				ports := containersPorts.GetIndex(i)
				if _, ok := ports.MustMap()["ports"]; !ok {
					continue
				}
				for j := range ports.MustArray() {
					port := ports.GetIndex(j)
					portName, err := port.Get("name").String()
					if err != nil {
						continue
					}
					k.podTargetPorts[portName] = port.Get("containerPort").MustInt()
				}
			}
			labels := metaData.Get("labels").MustMap()
			labelSlice := cloudcommon.StringInterfaceMapKVs(labels, ":")
			labelString := strings.Join(labelSlice, ", ")
			podGroup := model.PodGroup{
				Lcuuid:             uID,
				Name:               name,
				Label:              labelString,
				Type:               serviceType,
				PodNum:             replicas,
				PodNamespaceLcuuid: namespaceLcuuid,
				AZLcuuid:           k.azLcuuid,
				RegionLcuuid:       k.RegionUuid,
				PodClusterLcuuid:   common.GetUUID(k.UuidGenerate, uuid.Nil),
			}
			podGroups = append(podGroups, podGroup)
			k.podGroupLcuuids.Add(uID)
		}
	}
	log.Debug("get podgroups complete")
	return
}

func (k *KubernetesGather) getPodReplicationControllers() (podRCs []model.PodGroup, err error) {
	log.Debug("get replicationcontrollers starting")
	for _, r := range k.k8sInfo["*v1.ReplicationController"] {
		rData, rErr := simplejson.NewJson([]byte(r))
		if rErr != nil {
			err = rErr
			log.Errorf("replicationcontroller initialization simplejson error: (%s)", rErr.Error())
			return
		}
		metaData, ok := rData.CheckGet("metadata")
		if !ok {
			log.Info("replicationcontroller metadata not found")
			continue
		}
		uID := metaData.Get("uid").MustString()
		if uID == "" {
			log.Info("replicationcontroller uid not found")
			continue
		}
		name := metaData.Get("name").MustString()
		if name == "" {
			log.Infof("replicationcontroller (%s) name not found", uID)
			continue
		}
		namespace := metaData.Get("namespace").MustString()
		namespaceLcuuid, ok := k.namespaceToLcuuid[namespace]
		if !ok {
			log.Infof("replicationcontroller (%s) namespace not found", name)
			continue
		}
		label := "replicationcontroller:" + namespace + ":" + name
		serviceType := common.POD_GROUP_RC
		_, ok = k.nsLabelToGroupLcuuids[namespace+label]
		if ok {
			k.nsLabelToGroupLcuuids[namespace+label].Add(uID)
		} else {
			rcLcuuidsSet := mapset.NewSet()
			rcLcuuidsSet.Add(uID)
			k.nsLabelToGroupLcuuids[namespace+label] = rcLcuuidsSet
		}
		labels := rData.Get("spec").Get("selector").MustMap()
		for key, v := range labels {
			nsLabel := namespace + key + "_" + v.(string)
			_, ok := k.nsLabelToGroupLcuuids[nsLabel]
			if ok {
				k.nsLabelToGroupLcuuids[nsLabel].Add(uID)
			} else {
				nsRCLcuuidsSet := mapset.NewSet()
				nsRCLcuuidsSet.Add(uID)
				k.nsLabelToGroupLcuuids[nsLabel] = nsRCLcuuidsSet
			}
		}
		containersPorts := rData.Get("spec").Get("template").Get("spec").Get("containers")
		for i := range containersPorts.MustArray() {
			ports := containersPorts.GetIndex(i)
			if _, ok := ports.MustMap()["ports"]; !ok {
				continue
			}
			for j := range ports.MustArray() {
				port := ports.GetIndex(j)
				portName, err := port.Get("name").String()
				if err != nil {
					continue
				}
				k.podTargetPorts[portName] = port.Get("containerPort").MustInt()
			}
		}
		labelSlice := cloudcommon.StringInterfaceMapKVs(labels, ":")
		labelString := strings.Join(labelSlice, ",")
		podNum := rData.Get("spec").Get("replicas").MustInt()
		podRC := model.PodGroup{
			Lcuuid:             uID,
			Name:               name,
			Label:              labelString,
			Type:               serviceType,
			PodNum:             podNum,
			RegionLcuuid:       k.RegionUuid,
			AZLcuuid:           k.azLcuuid,
			PodNamespaceLcuuid: namespaceLcuuid,
			PodClusterLcuuid:   common.GetUUID(k.UuidGenerate, uuid.Nil),
		}
		podRCs = append(podRCs, podRC)
		k.podGroupLcuuids.Add(uID)
	}
	log.Debug("get replicationcontrollers complete")
	return
}
