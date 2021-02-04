// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package template

import (
	"bytes"
	"fmt"
	"reflect"
	tpl "text/template"

	"github.com/mitchellh/reflectwalk"
)

// walker struct implements the methods that reflectwalk will use in order to
// walk the tree and apply the template translation.  The walker struct must
// argument and the function definition (pointer receiver) must be the same.
// Because the methods on the struct are defined as pointer receiver, the
// following is how you would declare and use the walker (notice the pointer)
// to the walker.
//
//      w := &walker{
//		     tpl:    tpl,
//		     tree:   tree,
//		     values: values,
//	     }
//
//      err := reflectwalk.Walk(tree, w)
//
// To ensure this is done correctly, there is a helper walk() function
// that will construct the walker and run it against the template, tree,
// and values.
type walker struct {
	tpl    *tpl.Template
	tree   interface{}
	values interface{}
}

// walk will use the configured template object to apply values to the
// given tree.  It will modify the passed in tree.
func walk(tpl *tpl.Template, tree interface{}, values interface{}) error {
	w := &walker{
		tpl:    tpl,
		tree:   tree,
		values: values,
	}

	err := reflectwalk.Walk(tree, w)
	if err != nil {
		return fmt.Errorf("Could not walk tree: %w", err)
	}

	return nil
}

// ------ reflectWalk interface support -----------------------

// Primitive is able to handle all array elements.  In addition,
// if the value is a pointer to a string, this function can
// replace the value with a template translation.
func (w *walker) Primitive(v reflect.Value) error {
	//	fmt.Println("PRIMITIVE", v, ", KIND: ", v.Kind())

	if v.CanSet() {
		switch v.Kind() {
		case reflect.Interface:
			t, ok := v.Interface().(string)

			if ok {
				// fmt.Println("We have a string thing: ", t)
				newValue, err := w.applyTemplateValues(t)
				if err != nil {
					// TODO: Wrap error appropriately
					return fmt.Errorf("Failed applying values to primitive: value=%s: %w", t, err)
				}
				v.Set(reflect.ValueOf(newValue))
			}

		case reflect.String:
			newValue, err := w.applyTemplateValues(v.String())
			if err != nil {
				// TODO: Wrap error appropriately
				return fmt.Errorf("Failed applying values to primitive: value=%s: %w", v.String(), err)
			}

			v.Set(reflect.ValueOf(newValue))
		}
	}
	return nil
}

func (w *walker) Map(m reflect.Value) error {
	//	fmt.Println("MAP_STRUCTURE", m, ", CANSET: ", m.CanSet(), ", CANADDR", m.CanAddr())
	return nil
}

// MapElem supports the features of not only applying the template translation
// to a value, but it also is able to apply template logic to the keys itself.
// In order to do this, if the underlying object is a string we will run the
// translation engine on it.  We then delete the original map (by setting the
// index value to an empty value struct) and add in a new map key and value
// using the translated key.
func (w *walker) MapElem(m, k, v reflect.Value) error {
	// fmt.Println("MAP_ELEM_MAP", m, ", CANSET: ", m.CanSet(), ", CANADDR", m.CanAddr())
	// fmt.Println("MAP_ELEM_KEY", k, ", KIND: ", k.Kind())
	// fmt.Println("MAP_ELEM_VALUE", v, ", KIND: ", v.Kind(), ", CANSET: ", v.CanSet(), ", CANADDR", v.CanAddr())

	var valStr string
	switch v.Kind() {
	case reflect.Interface:
		t, ok := v.Interface().(string)
		if ok {
			valStr = t
		}
	case reflect.String:
		valStr = v.String()
	}

	newValue := v
	if valStr != "" {
		newValStr, err := w.applyTemplateValues(valStr)
		if err != nil {
			// TODO: Wrap error appropriately
			return fmt.Errorf("Failed applying values to map value: content=%s: %w", valStr, err)
		}
		newValue = reflect.ValueOf(newValStr)
	}

	// Support translation of the key itself
	var keyStr string
	switch k.Kind() {
	case reflect.Interface:
		t, ok := k.Interface().(string)
		if ok {
			keyStr = t
		}
	case reflect.String:
		keyStr = k.String()
	}

	newKey := k
	if keyStr != "" {
		newKeyStr, err := w.applyTemplateValues(keyStr)
		if err != nil {
			// TODO: Wrap error appropriately
			return fmt.Errorf("Failed applying values to map key: content=%s: %w", k.String(), err)
		}

		newKey = reflect.ValueOf(newKeyStr)
	}

	m.SetMapIndex(k, reflect.Value{})
	m.SetMapIndex(newKey, newValue)

	return nil
}

func (w *walker) applyTemplateValues(content string) (string, error) {

	var res bytes.Buffer

	tpl, err := w.tpl.Parse(content)
	if err != nil {
		return content, &ParseError{
			Err:     err,
			Content: content,
		}
	}

	err = tpl.Execute(&res, w.values)
	if err != nil {
		if isMissingKeyError(err) {
			return "", &MissingKeyError{
				Err:     err,
				Key:     getMissingKeyFromError(err),
				Content: content,
				Values:  w.values,
			}

		}

		return content, &ApplyValueError{
			Err:     err,
			Content: content,
			Values:  w.values,
		}
	}

	return res.String(), nil
}

// original MapElem implementation
// func (w *walker) MapElem(m, k, v reflect.Value) error {
// 	fmt.Println("MAP_ELEM_MAP", m, ", CANSET: ", m.CanSet(), ", CANADDR", m.CanAddr())
// 	fmt.Println("MAP_ELEM_KEY", k, ", KIND: ", k.Kind())
// 	fmt.Println("MAP_ELEM_VALUE", v, ", KIND: ", v.Kind(), ", CANSET: ", v.CanSet(), ", CANADDR", v.CanAddr())

// 	switch v.Kind() {

// 	case reflect.Interface:
// 		t, ok := v.Interface().(string)

// 		if ok {
// 			// fmt.Println("We have a string thing: ", t)
// 			newValue, err := w.applyTemplateValues(t)
// 			if err != nil {
// 				// TODO: Wrap error appropriately
// 				return errors.Wrapf(err, "Failed parsing of value %s", t)
// 			}

// 			m.SetMapIndex(k, reflect.ValueOf(newValue))

// 		}

// 	case reflect.String:
// 		newValue, err := w.applyTemplateValues(v.String())
// 		if err != nil {
// 			// TODO: Wrap error appropriately
// 			return errors.Wrapf(err, "Failed parsing of value %s", v.String())
// 		}

// 		m.SetMapIndex(k, reflect.ValueOf(newValue))

// 	}

// 	return nil
// }

// func (w *walker) Enter(l reflectwalk.Location) error {
// 	// fmt.Printf("Enter: %s\n", l.String())
// 	return nil
// }

// func (w *walker) Exit(l reflectwalk.Location) error {
// 	// fmt.Printf("Exit: %s\n", l.String())
// 	return nil
// }

// func (w *walker) Slice(v reflect.Value) error {
// 	return nil
// }

// func (w *walker) SliceElem(i int, v reflect.Value) error {
// 	// fmt.Println("SLICE_ELEM_VALUE", "INDEX: ", i, ", VALUE: ", v, ", KIND: ", v.Kind(), ", CANSET: ", v.CanSet(), ", CANADDR", v.CanAddr())

// 	switch v.Kind() {
// 	case reflect.Interface:
// 		t, ok := v.Interface().(string)

// 		if ok {
// 			// fmt.Println("We have a string thing: ", t)
// 			newValue, err := w.applyTemplateValues(t)
// 			if err != nil {
// 				// TODO: Wrap error appropriately
// 				return errors.Wrapf(err, "Failed parsing of value %s", t)
// 			}
// 			v.Set(reflect.ValueOf(newValue))
// 		}

// 	case reflect.String:
// 		newValue, err := w.applyTemplateValues(v.String())
// 		if err != nil {
// 			// TODO: Wrap error appropriately
// 			return errors.Wrapf(err, "Failed parsing of value %s", v.String())
// 		}

// 		v.Set(reflect.ValueOf(newValue))

// 	}

// 	return nil
// }
