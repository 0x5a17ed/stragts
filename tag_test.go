package stragts

import (
	"testing"

	assertpkg "github.com/stretchr/testify/assert"
)

func TestTag_Fill(t *testing.T) {
	assert := assertpkg.New(t)

	type TestStruct struct {
		NumField int

		BoolField bool

		StringField string

		Switch *bool

		Slice []string
	}

	v := &TestStruct{}

	tag := Tag{Value: "12,true,name,~switch"}
	assert.NoError(tag.Fill(v))

	assert.Equal(12, v.NumField)
	assert.True(v.BoolField)
	assert.Equal("name", v.StringField)

	assert.True(*v.Switch)

	tag = Tag{Value: "string-field='hello world'"}
	assert.NoError(tag.Fill(v))
	assert.Equal("hello world", v.StringField)

	tag = Tag{Value: "slice='foo';'baa'"}
	assert.NoError(tag.Fill(v))
	assert.Equal([]string{"foo", "baa"}, v.Slice)
}
