package lotgo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type exampleTest struct {
	Value int
}

func (t *exampleTest) SetUp(r *Runner) {
}

func (t *exampleTest) TearDown(r *Runner) {
}

func (t *exampleTest) Test(r *Runner) error {
	return nil
}

var _ LoadTest = &exampleTest{}

func TestLoadTest_deepClone(t *testing.T) {
	var lt LoadTest = &exampleTest{Value: 11}
	newlt := deepClone(lt)
	assert.False(t, lt == newlt)
	newtest := newlt.(*exampleTest)
	assert.Equal(t, 11, newtest.Value)
}

func TestLoadTest_AddTest(t *testing.T) {
	var lt LoadTest = &exampleTest{Value: 12}
	AddTest("foo", lt)
	AddTest("bar", lt)
	assert.Equal(t, "bar foo", AllTests())
}
