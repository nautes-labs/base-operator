// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"runtime/debug"

	"github.com/nautes-labs/base-operator/pkg/log"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func InArray(needle interface{}, haystack interface{}) bool {
	val := reflect.ValueOf(haystack)
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(needle, val.Index(i).Interface()) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			if reflect.DeepEqual(needle, val.MapIndex(k).Interface()) {
				return true
			}
		}
	default:
		panic("haystack: haystack type muset be slice, array or map")
	}

	return false
}

func JsonMarshalInterfaceToIOReader(data interface{}) (io.Reader, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not marshal data: %v", err)
	}

	return bytes.NewReader(b), nil
}

func GetObjectGvr(restMapper meta.RESTMapper, obj interface{}) schema.GroupVersionResource {
	emptyGvr := schema.GroupVersionResource{}
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return emptyGvr
	}
	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}
	gvk := unstructuredObj.GroupVersionKind()
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return emptyGvr
	}
	return schema.GroupVersionResource{
		Group:    mapping.Resource.Group,
		Version:  mapping.Resource.Version,
		Resource: mapping.Resource.Resource,
	}
}

func PanicTrace() {
	if err := recover(); err != nil {
		log.Loger.WithField("err_msg", err).WithField("trace_stack", string(debug.Stack())).Error("application occurrence of panic")
	}
}

func Intersect(a []string, b []string) []string {
	result := make([]string, 0)
	mp := make(map[string]struct{})

	for _, s := range a {
		if _, ok := mp[s]; !ok {
			mp[s] = struct{}{}
		}
	}
	for _, s := range b {
		if _, ok := mp[s]; ok {
			result = append(result, s)
		}
	}
	return result
}

func DiffArray(a []string, b []string) []string {
	var diffArray []string
	temp := map[string]struct{}{}

	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}

func DeleteArrayItem(item string, list []string) {
	index := 0
	for i, v := range list {
		if v == item {
			index = i
		}
	}
	list = append(list[:index], list[index+1:]...)
}
