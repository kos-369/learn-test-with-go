package reflection

import "reflect"

type pointerVisit struct {
	typ     reflect.Type
	address uintptr
}

func walk(x any, fn func(input string)) {
	onPath := make(map[pointerVisit]struct{})

	var walkValue func(reflect.Value)

	walkValue = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}

		if value.Kind() == reflect.Interface {
			if value.IsNil() {
				return
			}

			walkValue(value.Elem())
			return
		}

		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return
			}

			current := pointerVisit{
				typ:     value.Type(),
				address: value.Pointer(),
			}

			if _, found := onPath[current]; found {
				return
			}

			onPath[current] = struct{}{}

			defer delete(onPath, current)

			walkValue(value.Elem())
			return
		}

		switch value.Kind() {
		case reflect.String:
			fn(value.String())

		case reflect.Struct:
			for i := 0; i < value.NumField(); i++ {
				walkValue(value.Field(i))
			}

		case reflect.Slice, reflect.Array:
			for i := 0; i < value.Len(); i++ {
				walkValue(value.Index(i))
			}

		case reflect.Map:
			for _, key := range value.MapKeys() {
				walkValue(value.MapIndex(key))
			}

		case reflect.Chan:
			for {
				received, ok := value.Recv()
				if !ok {
					return
				}

				walkValue(received)
			}

		case reflect.Func:
			if value.IsNil() || value.Type().NumIn() != 0 {
				return
			}

			results := value.Call(nil)
			for _, result := range results {
				walkValue(result)
			}
		}
	}

	walkValue(reflect.ValueOf(x))
}
