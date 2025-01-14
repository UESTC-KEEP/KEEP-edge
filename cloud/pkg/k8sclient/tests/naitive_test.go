package tests

import (
	"bytes"
	"context"
	"fmt"
	"github.com/UESTC-KEEP/keep/cloud/pkg/k8sclient/config"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os/user"
	"testing"
)

const metaCRD = `
apiVersion: "apiextensions.k8s.io/v1beta1"
kind: "CustomResourceDefinition"
metadata:
  name: "projects.example.lnf.com"
spec:
  group: "example.lnf.com"
  version: "v1alpha1"
  scope: "Namespaced"
  names:
    plural: "projects"
    singular: "project"
    kind: "Project"
  validation:
    openAPIV3Schema:
      required: ["spec"]
      properties:
        spec:
          required: ["replicas"]
          properties:
            replicas:
              type: "integer"
              minimum: 1
`

func TestCreatResourcesByYAML(t *testing.T) {
	config.GetClient()
	//CreatResourcesByYAML(constants.DefaultRedisConfigMapConfigMap, "default")
}

func TestOpCRD(t *testing.T) {
	obj := &unstructured.Unstructured{}
	_, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode([]byte(metaCRD), nil, obj)
	if err != nil {
		log.Fatal(err)
	}
	dr, err := config.GetGVRdyClient(gvk, obj.GetNamespace())
	//if err!=nil{
	//	panic(fmt.Errorf("failed to get dr: %v",err))
	//}
	//objCreate, err := dr.Create(context.TODO(),obj, metav1.CreateOptions{})
	//if err!=nil{
	//	panic(fmt.Errorf("failed to createCrd"))
	//}
	//log.Print("Create: : ",objCreate)
	unstructuredList, err := dr.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
	content := unstructuredList.UnstructuredContent()
	CRDList := &v1.PodList{}
	runtime.DefaultUnstructuredConverter.FromUnstructured(content, CRDList)
	for _, item := range CRDList.Items {
		t.Logf("%v\t %v\t %v\n",
			item.Namespace,
			item.Status.Phase,
			item.Name)
	}
}

func TestCreation(t *testing.T) {
	home := GetHomePath()
	nameSpace := "demo01"
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home)) // 使用 kubectl 默认配置 ~/.kube/config
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	// 创建一个k8s客户端
	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	dd, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatal(err)
	}

	filebytes, err := ioutil.ReadFile("/etc/keepedge/ymls/redis-standalone-conf.yaml")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(filebytes), 100)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			log.Fatal(err)
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		gr, err := restmapper.GetAPIGroupResources(clientSet.Discovery())
		if err != nil {
			log.Fatal(err)
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			log.Fatal(err)
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace(nameSpace)
			}
			dri = dd.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = dd.Resource(mapping.Resource)
		}

		obj2, err := dri.Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s/%s created", obj2.GetKind(), obj2.GetName())
	}
}

func GetHomePath() string {
	u, err := user.Current()
	if err == nil {
		return u.HomeDir
	}
	return ""
}
