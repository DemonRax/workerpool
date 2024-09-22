package worker_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	worker "github.com/freiheit-com/dentsply-sirona-lightning-core/restore/worker_pool"
)

type testData struct {
	msg string
	err error
}

func Test_Worker_Pool(t *testing.T) {
	testErr := errors.New("a < 0")
	wp := worker.New(5,
		func(raw any) error {
			data := raw.(testData)
			return data.err
		})

	wp.Start()

	for i, a := range []int{12, -5, 5, 18, -15, 4, -8, 19, 54} {
		data := testData{msg: fmt.Sprintf("%d %d", i, a)}
		if a < 0 {
			data.err = testErr
		}
		wp.Add(data)
	}

	wantErrs := worker.Errs{testErr, testErr, testErr}

	if diff := cmp.Diff(wantErrs, wp.Wait(), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("did not get expected result (-want/+got):\n%s", diff)
	}
	if want := 6; wp.Success() != want {
		t.Errorf("wp.Success() = %d, want %d", wp.Success(), want)
	}
}
