package cache

import "testing"

func TestVeryFragmentedFile(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	var file = "long"
	app.fsmap[file] = generateMeAString(1, 2000)
	defer app.cleanup()

	app.testRange(file, 5, 10)
	app.testRange(file, 5, 2)
	app.testRange(file, 2, 2)
	app.testRange(file, 20, 10)
	app.testRange(file, 30, 10)
	app.testRange(file, 40, 10)
	app.testRange(file, 50, 10)
	app.testRange(file, 60, 10)
	app.testRange(file, 70, 10)
	app.testRange(file, 50, 20)
	app.testRange(file, 200, 5)
	app.testFullRequest(file)
	app.testRange(file, 3, 1000)
}

func Test2PartsFile(t *testing.T) {
	var fsmap = make(map[string]string)
	var file = "2parts"
	fsmap[file] = generateMeAString(2, 10)
	t.Parallel()
	app := newTestAppFromMap(t, fsmap)
	defer app.cleanup()
	app.testRange(file, 2, 8)
	app.testFullRequest(file)
}
