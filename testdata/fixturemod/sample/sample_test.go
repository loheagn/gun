package sample

import (
	"fmt"
	"strconv"
	"testing"
)

const suffixConst = "const-sub"
const fmtID = 7

// marker:outside
func TestAlpha(t *testing.T) {
	t.Run("outer", func(t *testing.T) {
		t.Run("inner", func(t *testing.T) {
			t.Log("RUN:Alpha/outer/inner") // marker:inner
		})

		dynamicName := "dyn"
		t.Run(dynamicName, func(t *testing.T) {
			t.Log("RUN:Alpha/dyn") // marker:dynamic
		})
	})
}

func TestConst(t *testing.T) {
	t.Run("prefix-"+suffixConst, func(t *testing.T) {
		t.Log("RUN:Const/prefix-const-sub") // marker:const_concat
	})

	t.Run(fmt.Sprintf("fmt-%d", fmtID), func(t *testing.T) {
		t.Log("RUN:Const/fmt-7") // marker:fmt_sprintf
	})

	t.Run(strconv.Itoa(11), func(t *testing.T) {
		t.Log("RUN:Const/11") // marker:itoa
	})
}

func TestNoSub(t *testing.T) {
	t.Log("RUN:NoSub") // marker:nosub
}

func TestBeta(t *testing.T) {
	t.Log("RUN:Beta") // marker:beta
}
