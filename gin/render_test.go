package gin_test

import (
	"context"
	"testing"

	"github.com/ccmonky/inithook"
)

func init() {
	err := inithook.ExecuteAttrSetters(context.Background(), inithook.AppName, "myapp")
	if err != nil {
		panic(err)
	}
	err = inithook.ExecuteAttrSetters(context.Background(), inithook.Version, "0.3.0")
	if err != nil {
		panic(err)
	}
}

func TestGin(t *testing.T) {

}
