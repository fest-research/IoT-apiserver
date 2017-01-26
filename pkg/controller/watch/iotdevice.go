package watch

import (
	"log"

	"strconv"

	types "github.com/fest-research/iot-addon/pkg/api/v1"
	"github.com/fest-research/iot-addon/pkg/common"
	"github.com/fest-research/iot-addon/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type IotDeviceWatcher struct {
	dynamicClient *dynamic.Client
	restClient    *rest.RESTClient
}

var iotDeviceResource = metav1.APIResource{
	Name:       types.IotDeviceType,
	Namespaced: true,
}

func NewIotDeviceWatcher(dynamicClient *dynamic.Client, restClient *rest.RESTClient) IotDeviceWatcher {
	return IotDeviceWatcher{dynamicClient: dynamicClient, restClient: restClient}
}

func (w IotDeviceWatcher) Watch() {
	watcher, err := w.dynamicClient.
		Resource(&iotDeviceResource, api.NamespaceAll).
		Watch(&api.ListOptions{})

	if err != nil {
		log.Println(err.Error())
	}

	defer watcher.Stop()

	for {
		e, ok := <-watcher.ResultChan()

		if !ok {
			panic("IotDevices ended early?")
		}

		iotDevice, _ := e.Object.(*types.IotDevice)

		if e.Type == watch.Added || e.Type == watch.Modified {
			log.Printf("Device added %s\n", iotDevice.Metadata.Name)
			err := w.addModifyDeviceHandler(*iotDevice)
			if err != nil {
				log.Printf("Error [addModifyDeviceHandler] %s", err.Error())
			}
		} else if e.Type == watch.Error {
			log.Println("Error")
			break
		}
	}
}

func (w IotDeviceWatcher) addModifyDeviceHandler(iotDevice types.IotDevice) error {

	unschedulable := GetUnschedulableLabelFromDevice(iotDevice)
	deviceName := iotDevice.Metadata.Name

	if unschedulable {
		log.Printf("[addModifyDeviceHandler] Delete pods for unschedulable device %s", deviceName)
		pods, err := kubernetes.GetDevicePods(w.restClient, iotDevice)
		if err != nil {
			return err
		}

		for _, pod := range pods {
			err := w.deletePod(pod)
			if err != nil {
				return err
			}
		}

	} else {
		daemonSets, _ := kubernetes.GetDeviceDaemonSets(w.restClient, iotDevice)
		for _, ds := range daemonSets {

			if !kubernetes.IsPodCreated(w.restClient, ds, iotDevice) {
				log.Printf("[addModifyDeviceHandler] Create new pod %s ", ds.Metadata.Name)
				err := w.createPod(ds, deviceName)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (w IotDeviceWatcher) createPod(ds types.IotDaemonSet, deviceName string) error {

	newPod := types.IotPod{}
	name := ds.Metadata.Name

	pod := types.IotPod{
		TypeMeta: createTypeMeta(ds.APIVersion),
		Metadata: v1.ObjectMeta{
			Name:      name + "-" + string(common.NewUUID()),
			Namespace: ds.Metadata.Namespace,
			Labels: map[string]string{
				types.CreatedBy:      types.IotDaemonSetType + "." + name,
				types.DeviceSelector: deviceName,
			},
		},
		Spec: ds.Spec.Template.Spec,
	}

	return w.restClient.Post().
		Namespace(pod.Metadata.Namespace).
		Resource(types.IotPodType).
		Body(&pod).
		Do().
		Into(&newPod)

}

func (w IotDeviceWatcher) deletePod(pod types.IotPod) error {

	return w.restClient.Delete().
		Namespace(pod.Metadata.Namespace).
		Resource(types.IotPodType).
		Name(pod.Metadata.Name).
		Body(&v1.DeleteOptions{}).
		Do().
		Error()

}

func createTypeMeta(apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       types.IotPodKind,
		APIVersion: apiVersion,
	}
}

func GetUnschedulableLabelFromDevice(iotDevice types.IotDevice) bool {

	unschedulableLabel, ok := iotDevice.Metadata.Labels[types.Unschedulable]

	if ok {
		unschedulable, err := strconv.ParseBool(unschedulableLabel)
		if err != nil {
			return false
		}
		return unschedulable
	}
	return false
}
