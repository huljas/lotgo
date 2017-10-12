package lotgo

import (
	"reflect"
	"sort"
	"strings"
)

type LoadTest interface {
	/* Called before the actual test run, once per every thread. */
	SetUp(lt *Runner)

	/* Called after the test run, once per every thread. */
	TearDown(lt *Runner)

	/* Called for every test pass. */
	Test(lt *Runner) error
}

func deepClone(test LoadTest) LoadTest {
	t := reflect.TypeOf(test).Elem()
	vold := reflect.ValueOf(test)
	v := reflect.New(t)
	count := v.Elem().NumField()
	for i := 0; i < count; i++ {
		if v.Elem().Field(i).CanSet() {
			v.Elem().Field(i).Set(reflect.ValueOf(vold.Elem().Field(i).Interface()))
		}
	}
	return v.Interface().(LoadTest)
}


var regTests map[string]LoadTest = map[string]LoadTest{}

func GetTest(name string) (LoadTest, bool) {
	v, ok := regTests[name]
	return v, ok
}

func AddTest(name string, test LoadTest) LoadTest {
	regTests[name] = test
	return test
}

func AllTests() string {
	var keys []string
	for key := range regTests {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, " ")
}
