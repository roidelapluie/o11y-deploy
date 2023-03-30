// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modules

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

const (
	configFieldPrefix = "AUTO_MODULE_"
)

var (
	configNames      = make(map[string]Config)
	configFieldNames = make(map[reflect.Type]string)
	configFields     []reflect.StructField

	configTypesMu sync.Mutex
	configTypes   = make(map[reflect.Type]reflect.Type)

	emptyStructType = reflect.TypeOf(struct{}{})
	configsType     = reflect.TypeOf(Configs{})
)

// RegisterConfig registers the given Config type for YAML marshaling and unmarshaling.
func RegisterConfig(config Config) {
	registerConfig(config.Name()+"_module", reflect.TypeOf(config), config)
}

func registerConfig(yamlKey string, elemType reflect.Type, config Config) {
	name := config.Name()
	if _, ok := configNames[name]; ok {
		panic(fmt.Sprintf("module: Config named %q is already registered", name))
	}
	configNames[name] = config

	fieldName := configFieldPrefix + yamlKey // Field must be exported.
	configFieldNames[elemType] = fieldName

	// Insert fields in sorted order.
	i := sort.Search(len(configFields), func(k int) bool {
		return fieldName < configFields[k].Name
	})
	configFields = append(configFields, reflect.StructField{}) // Add empty field at end.
	copy(configFields[i+1:], configFields[i:])                 // Shift fields to the right.
	configFields[i] = reflect.StructField{                     // Write new field in place.
		Name: fieldName,
		Type: elemType,
		Tag:  reflect.StructTag(`yaml:"` + yamlKey + `,omitempty"`),
	}
}

func getConfigType(out reflect.Type) reflect.Type {
	configTypesMu.Lock()
	defer configTypesMu.Unlock()
	if typ, ok := configTypes[out]; ok {
		return typ
	}
	// Initial exported fields map one-to-one.
	var fields []reflect.StructField
	for i, n := 0, out.NumField(); i < n; i++ {
		switch field := out.Field(i); {
		case field.PkgPath == "" && field.Type != configsType:
			fields = append(fields, field)
		default:
			fields = append(fields, reflect.StructField{
				Name:    "_" + field.Name, // Field must be unexported.
				PkgPath: out.PkgPath(),
				Type:    emptyStructType,
			})
		}
	}
	// Append extra config fields on the end.
	fields = append(fields, configFields...)
	typ := reflect.StructOf(fields)
	configTypes[out] = typ
	return typ
}

// UnmarshalYAMLWithInlineConfigs helps implement yaml.Unmarshal for structs
// that have a Configs field that should be inlined.
func UnmarshalYAMLWithInlineConfigs(out interface{}, unmarshal func(interface{}) error) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("module: can only unmarshal into a struct pointer: %T", out)
	}
	outVal = outVal.Elem()
	if outVal.Kind() != reflect.Struct {
		return fmt.Errorf("module: can only unmarshal into a struct pointer: %T", out)
	}
	outTyp := outVal.Type()

	cfgTyp := getConfigType(outTyp)
	cfgPtr := reflect.New(cfgTyp)
	cfgVal := cfgPtr.Elem()

	// Copy shared fields (defaults) to dynamic value.
	var configs *Configs
	for i, n := 0, outVal.NumField(); i < n; i++ {
		if outTyp.Field(i).Type == configsType {
			configs = outVal.Field(i).Addr().Interface().(*Configs)
			continue
		}
		if cfgTyp.Field(i).PkgPath != "" {
			continue // Field is unexported: ignore.
		}
		cfgVal.Field(i).Set(outVal.Field(i))
	}
	if configs == nil {
		return fmt.Errorf("module: Configs field not found in type: %T", out)
	}

	// Unmarshal into dynamic value.
	if err := unmarshal(cfgPtr.Interface()); err != nil {
		return replaceYAMLTypeError(err, cfgTyp, outTyp)
	}

	// Copy shared fields from dynamic value.
	for i, n := 0, outVal.NumField(); i < n; i++ {
		if cfgTyp.Field(i).PkgPath != "" {
			continue // Field is unexported: ignore.
		}
		outVal.Field(i).Set(cfgVal.Field(i))
	}

	var err error
	*configs, err = readConfigs(cfgVal, outVal.NumField())
	return err
}

func readConfigs(structVal reflect.Value, startField int) (Configs, error) {
	var (
		configs Configs
	)
	for i, n := startField, structVal.NumField(); i < n; i++ {
		field := structVal.Field(i)
		if field.Kind() != reflect.Ptr {
			panic("module: internal error: field is not a pointer")
		}
		var val reflect.Value
		if field.IsNil() {
			val = reflect.New(field.Type().Elem()) // Get the zero value of the field's type
			// Manually call UnmarshalYAML on the zero value
			if unmarshaler, ok := val.Interface().(yaml.Unmarshaler); ok {
				err := unmarshaler.UnmarshalYAML(func(interface{}) error { return nil })
				if err != nil {
					return nil, err
				}
			}
		} else {
			val = field
		}

		c, ok := val.Interface().(Config)
		if !ok {
			panic("module: internal error: field element is not a Config")
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// MarshalYAMLWithInlineConfigs helps implement yaml.Marshal for structs
// that have a Configs field that should be inlined.
func MarshalYAMLWithInlineConfigs(in interface{}) (interface{}, error) {
	inVal := reflect.ValueOf(in)
	for inVal.Kind() == reflect.Ptr {
		inVal = inVal.Elem()
	}
	inTyp := inVal.Type()

	cfgTyp := getConfigType(inTyp)
	cfgPtr := reflect.New(cfgTyp)
	cfgVal := cfgPtr.Elem()

	// Copy shared fields to dynamic value.
	var configs *Configs
	for i, n := 0, inTyp.NumField(); i < n; i++ {
		if inTyp.Field(i).Type == configsType {
			configs = inVal.Field(i).Addr().Interface().(*Configs)
		}
		if cfgTyp.Field(i).PkgPath != "" {
			continue // Field is unexported: ignore.
		}
		cfgVal.Field(i).Set(inVal.Field(i))
	}
	if configs == nil {
		return nil, fmt.Errorf("module: Configs field not found in type: %T", in)
	}

	if err := writeConfigs(cfgVal, *configs); err != nil {
		return nil, err
	}

	return cfgPtr.Interface(), nil
}

func writeConfigs(structVal reflect.Value, configs Configs) error {
	for _, c := range configs {
		fieldName, ok := configFieldNames[reflect.TypeOf(c)]
		if !ok {
			return fmt.Errorf("module: cannot marshal unregistered Config type: %T", c)
		}
		field := structVal.FieldByName(fieldName)
		field.Set(reflect.ValueOf(c))
	}
	return nil
}

func replaceYAMLTypeError(err error, oldTyp, newTyp reflect.Type) error {
	var e *yaml.TypeError
	if errors.As(err, &e) {
		oldStr := oldTyp.String()
		newStr := newTyp.String()
		for i, s := range e.Errors {
			e.Errors[i] = strings.Replace(s, oldStr, newStr, -1)
		}
	}
	return err
}
