package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

/*
This test checks that you in fact implemented a pipeline and not accumulating all results (until the input
is exhausted) and only then passing it to next function in pipeline. This is a wrong approach! It will not
allow to work with tasks that requires input of undefined length. You need to pass the computed result of
the function to the next function in pipeline as soon as it is ready.
*/
func TestPipeline(t *testing.T) {

	var ok = true
	var recieved uint32
	freeFlowJobs := []job{
		job(func(in, out chan any) {
			out <- 1
			time.Sleep(10 * time.Millisecond)
			currRecieved := atomic.LoadUint32(&recieved)

			// Here is a gist of this test:
			// If you are accumulating values instead of implementing a pipeline, you are not passing values to the
			// next funcion before your current function is finished. That is what I am checking: counter
			// should increase in next function (meaning that values are going there) before current function
			// finished its execution.

			if currRecieved == 0 {
				ok = false
			}
		}),
		job(func(in, out chan any) {
			for range in {
				atomic.AddUint32(&recieved, 1)
			}
		}),
	}
	ExecutePipeline(freeFlowJobs...)
	if !ok || recieved == 0 {
		t.Errorf("no value free flow - dont collect them")
	}
}

func TestSigner(t *testing.T) {

	testExpected := "1173136728138862632818075107442090076184424490584241521304_1696913515191343735512658979631549563179965036907783101867_27225454331033649287118297354036464389062965355426795162684_29568666068035183841425683795340791879727309630931025356555_3994492081516972096677631278379039212655368881548151736_4958044192186797981418233587017209679042592862002427381542_4958044192186797981418233587017209679042592862002427381542"
	testResult := "NOT_SET"

	// This is a small check to verify that you are actually using supplied `DataSignerMd5` and
	// `DataSignerCrc32` functions. These function are substituted by the ones that are incrementing
	// some local counter. Substitution is possible due to the fact that functions are passed as
	// variables.

	var (
		DataSignerSalt         string = "" // на сервере будет другое значение
		OverheatLockCounter    uint32
		OverheatUnlockCounter  uint32
		DataSignerMd5Counter   uint32
		DataSignerCrc32Counter uint32
	)
	OverheatLock = func() {
		atomic.AddUint32(&OverheatLockCounter, 1)
		for {
			if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
				fmt.Println("OverheatLock happend")
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	OverheatUnlock = func() {
		atomic.AddUint32(&OverheatUnlockCounter, 1)
		for {
			if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
				fmt.Println("OverheatUnlock happend")
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	DataSignerMd5 = func(data string) string {
		atomic.AddUint32(&DataSignerMd5Counter, 1)
		OverheatLock()
		defer OverheatUnlock()
		data += DataSignerSalt
		dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
		time.Sleep(10 * time.Millisecond)
		return dataHash
	}
	DataSignerCrc32 = func(data string) string {
		atomic.AddUint32(&DataSignerCrc32Counter, 1)
		data += DataSignerSalt
		crcH := crc32.ChecksumIEEE([]byte(data))
		dataHash := strconv.FormatUint(uint64(crcH), 10)
		time.Sleep(time.Second)
		return dataHash
	}

	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	hashSignJobs := []job{
		job(func(in, out chan any) {
			for _, num := range inputData {
				out <- num
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan any) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				t.Error("cant convert result data to string")
			}
			testResult = data
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)

	end := time.Since(start)

	expectedTime := 3 * time.Second

	if testExpected != testResult {
		t.Errorf("results dont match\nGot: %v\nExpected: %v", testResult, testExpected)
	}

	if end > expectedTime {
		t.Errorf("execution took too long\nGot: %s\nExpected: <%s", end, time.Second*3)
	}

	// 2 calls in SingleHash and 6 calls in MultiHash
	if int(OverheatLockCounter) != len(inputData) ||
		int(OverheatUnlockCounter) != len(inputData) ||
		int(DataSignerMd5Counter) != len(inputData) ||
		int(DataSignerCrc32Counter) != len(inputData)*8 {
		t.Errorf("not enough hash-func calls")
	}

}

func TestByIlia(t *testing.T) {

	var recieved uint32
	freeFlowJobs := []job{
		job(func(in, out chan any) {
			out <- uint32(1)
			out <- uint32(3)
			out <- uint32(4)
		}),
		job(func(in, out chan any) {
			for val := range in {
				out <- val.(uint32) * 3
				time.Sleep(time.Millisecond * 100)
			}
		}),
		job(func(in, out chan any) {
			for val := range in {
				fmt.Println("collected", val)
				atomic.AddUint32(&recieved, val.(uint32))
			}
		}),
	}

	start := time.Now()

	ExecutePipeline(freeFlowJobs...)

	end := time.Since(start)

	expectedTime := time.Millisecond * 350

	if end > expectedTime {
		t.Errorf("execution took too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}

	if recieved != (1+3+4)*3 {
		t.Errorf("f3 have not collected inputs, recieved = %d", recieved)
	}
}
