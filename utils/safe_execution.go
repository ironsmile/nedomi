package utils

// SafeExecute executes f and recovers from any panics inside it
// if panic is found it will be wrapped with NewErrorWithStack and used with the
// provided print func
func SafeExecute(f func(), print func(error)) {
	defer func() {
		if str := recover(); str != nil {
			print(NewErrorWithStack(str.(string)))
		}
	}()
	f()
}
