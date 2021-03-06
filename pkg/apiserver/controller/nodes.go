package controller

import (
	"github.com/fest-research/iot-addon/pkg/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/watch"

	kubeapi "k8s.io/client-go/pkg/api/v1"
)

type INodeController interface {
	// TransformWatchEvent implements WatchEventController.
	TransformWatchEvent(event watch.Event) watch.Event
	ToNodeList(*v1.IotDeviceList) *kubeapi.NodeList
	ToNode(*v1.IotDevice) *kubeapi.Node
	ToIotDevice(*kubeapi.Node) *v1.IotDevice
	ToUnstructured(*kubeapi.Node) (*unstructured.Unstructured, error)
	ToBytes(*unstructured.Unstructured) ([]byte, error)
}

type nodeController struct {
	iotDomain string
}

// TransformWatchEvent converts an ADD/UPDATE/DELETE event for an IotDevice to
// an ADD/UPDATE/DELETE event for a k8s Node
func (this nodeController) TransformWatchEvent(event watch.Event) watch.Event {
	iotDevice := event.Object.(*v1.IotDevice)
	event.Object = this.ToNode(iotDevice)
	return event
}

// ToNodeList converts a list of IotDevices to a list of k8s Nodes
func (this nodeController) ToNodeList(iotDeviceList *v1.IotDeviceList) *kubeapi.NodeList {
	nodeList := &kubeapi.NodeList{}

	nodeList.TypeMeta = this.getTypeMeta(v1.NodeListKind)
	nodeList.Items = make([]kubeapi.Node, 0)

	for _, iotDevice := range iotDeviceList.Items {
		node := this.ToNode(&iotDevice)
		nodeList.Items = append(nodeList.Items, *node)
	}

	return nodeList
}

// ToNode converts an IotDevice object to a k8s Node object
func (this nodeController) ToNode(iotDevice *v1.IotDevice) *kubeapi.Node {
	node := &kubeapi.Node{}

	// TODO: subject to revision
	node.TypeMeta = this.getTypeMeta(v1.NodeKind)

	node.Spec = iotDevice.Spec
	node.Status = iotDevice.Status
	node.ObjectMeta = iotDevice.Metadata

	node.ObjectMeta.Namespace = ""

	return node
}

// ToIotDevice converts a k8s Node object to an IotDevice object
func (this nodeController) ToIotDevice(node *kubeapi.Node) *v1.IotDevice {
	iotDevice := &v1.IotDevice{}

	// TODO: subject to revision
	iotDevice.TypeMeta = this.getIotTypeMeta()

	iotDevice.Metadata = node.ObjectMeta
	iotDevice.Status = node.Status
	iotDevice.Spec = node.Spec

	// TODO: should we set namespace of iot device? Get from DB?

	return iotDevice
}

// ToUnstructured converts node to unstructured iot device
func (this nodeController) ToUnstructured(node *kubeapi.Node) (*unstructured.Unstructured, error) {
	result := &unstructured.Unstructured{}
	iotDevice := this.ToIotDevice(node)

	marshalledIotDevice, err := json.Marshal(iotDevice)
	if err != nil {
		return nil, err
	}

	err = result.UnmarshalJSON(marshalledIotDevice)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ToBytes converts unstructured iot device to node json bytes array
func (this nodeController) ToBytes(unstructured *unstructured.Unstructured) ([]byte, error) {
	marshalledIotDevice, err := unstructured.MarshalJSON()
	if err != nil {
		return nil, err
	}

	iotDevice := &v1.IotDevice{}
	err = json.Unmarshal(marshalledIotDevice, iotDevice)
	if err != nil {
		return nil, err
	}

	node := this.ToNode(iotDevice)
	marshalledNode, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}

	return marshalledNode, nil
}

func (this nodeController) getIotTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: this.iotDomain + "/" + v1.APIVersion,
		Kind:       v1.IotDeviceKind,
	}
}

func (this nodeController) getTypeMeta(kind v1.ResourceKind) metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: v1.APIVersion,
		Kind:       string(kind),
	}
}

func NewNodeController(iotDomain string) INodeController {
	return &nodeController{iotDomain: iotDomain}
}
